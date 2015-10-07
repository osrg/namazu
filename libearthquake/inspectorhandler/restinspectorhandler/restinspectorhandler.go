// Copyright (C) 2015 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restinspectorhandler

import (
	. "../../equtils"
	"encoding/json"
	"fmt"
	. "github.com/ahmetalpbalkan/go-linq"
	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type RESTInspectorHandler struct {
}

type RESTEntity struct {
	ActionsLock       sync.RWMutex
	Actions           []*Action
	HttpActionReadyCh chan bool
	Entity            TransitionEntity
}

//TODO: move global vars to RESTInspectorHandler struct
var (
	RESTEntities  = map[string]*RESTEntity{}
	ReadyEntityCh chan *TransitionEntity
)

func isEntityRegistered(entityId string) bool {
	_, alreadyRegistered := RESTEntities[entityId]
	return alreadyRegistered
}

func registerEntity(entityId string) error {
	if isEntityRegistered(entityId) {
		return fmt.Errorf("entity %s has been already registered", entityId)
	}
	restEntity := RESTEntity{
		ActionsLock:       sync.RWMutex{},
		Actions:           make([]*Action, 0),
		HttpActionReadyCh: make(chan bool),
		Entity: TransitionEntity{
			Id:             entityId,
			Conn:           nil,
			ActionFromMain: make(chan *Action),
			EventToMain:    make(chan *Event),
		},
	}
	RESTEntities[entityId] = &restEntity
	log.Debugf("Registered entity: %s", entityId)
	go func() {
		for {
			// TODO: how to shutdown this goroutine
			act := <-restEntity.Entity.ActionFromMain
			log.Debugf("(From Main)Received action: %s", act)
			restEntity.ActionsLock.Lock()
			restEntity.Actions = append(restEntity.Actions, act)
			restEntity.ActionsLock.Unlock()
			restEntity.HttpActionReadyCh <- true
		}
	}()
	return nil
}

// @app.route('/', methods=['GET'])
func RootOnGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello Earthquake! -- RESTInspectorHandler(pid=%d)\n", os.Getpid())
}

func makeEventStruct(entityId string, reader io.Reader) (ev Event, err error) {
	var m map[string]interface{}
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&m)
	if err != nil {
		return
	}
	ev, err = EventFromJSONMap(m, time.Now(), entityId)
	return
}

// @app.route(api_root + '/events/<entity_id>/<event_uuid>', methods=['POST'])
func EventsOnPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityId := vars["entity_id"]
	eventUuid := vars["event_uuid"]
	log.Debugf("EventsOnPost: entityId=%s, eventUuid=%s", entityId, eventUuid)
	if !isEntityRegistered(entityId) {
		if err := registerEntity(entityId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// parse event
	e, err := makeEventStruct(entityId, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send event to entity
	entity := &RESTEntities[entityId].Entity
	go func() {
		log.Debugf("*EventsOnPost: sending %s to %s", e, entityId)
		entity.EventToMain <- &e
		log.Debugf("*EventsOnPost: sent %s to %s", e, entityId)

	}()

	go func() {
		log.Debugf("*EventsOnPost: notifying %s to main", entityId)
		ReadyEntityCh <- entity
		log.Debugf("*EventsOnPost: notified %s to main", entityId)
	}()

	// return empty json
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{}")
}

func (this *RESTEntity) WaitForAction() (act *Action) {
	for {
		this.ActionsLock.RLock()
		l := len(this.Actions)
		this.ActionsLock.RUnlock()
		if l > 0 {
			break
		}
		log.Debugf("Waiting for action (entityId=%s)", this.Entity.Id)
		<-this.HttpActionReadyCh
	}
	this.ActionsLock.RLock()
	act = this.Actions[0]
	this.ActionsLock.RUnlock()
	if act.ActionType != "_JSON" {
		panic(log.Criticalf("Invalid action type %s", act.ActionType))
	}
	return
}

// @app.route(api_root + '/actions/<entity_id>', methods=['GET'])
func ActionsOnGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityId := vars["entity_id"]
	log.Debugf("ActionsOnGet: entityId=%s", entityId)
	if !isEntityRegistered(entityId) {
		if err := registerEntity(entityId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// TODO: refactor this logic
	ch := make(chan *Action)
	go func() {
		ch <- RESTEntities[entityId].WaitForAction()
	}()
	act := <-ch

	log.Debugf("Action %s ready for entityId=%s", act, entityId)
	// NOTE: an inspector has responsibility to DELETE the action,
	//       because HTTP DELETE must be idempotent (RFC 7231)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(act.ToJSONMap()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// @app.route(api_root + '/actions/<entity_id>/<action_uuid>', methods=['DELETE'])
func ActionsOnDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityId := vars["entity_id"]
	actionUuid := vars["action_uuid"]
	log.Debugf("ActionsOnDelete: entityId=%s, actionUuid=%s", entityId, actionUuid)
	if !isEntityRegistered(entityId) {
		err := fmt.Errorf("Unknown entity %s", entityId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	restEntity := RESTEntities[entityId]
	restEntity.ActionsLock.Lock()
	defer restEntity.ActionsLock.Unlock()

	// DELETE the action
	// NOTE: newActions == Actions is expected, as DELETE must be idempotent (RFC 7231)
	newActions, err := From(restEntity.Actions).Where(func(s T) (bool, error) {
		return s.(*Action).ToJSONMap()["uuid"].(string) != actionUuid, nil
	}).Results()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	deletedActions := len(restEntity.Actions) - len(newActions)
	log.Debugf("Deleted %d actions", deletedActions)
	if !(deletedActions == 0 || deletedActions == 1) {
		panic(log.Criticalf("this should not happen. deletedActions=%d", deletedActions))
	}

	// this fails:
	//  restEntity.Actions = newActions.([]*Action)
	//  --> invalid type assertion: newActions.([]*equtils.Action) (non-interface type []linq.T on left)
	restEntity.Actions = make([]*Action, len(newActions))
	for i, _ := range newActions {
		restEntity.Actions[i] = newActions[i].(*Action)
	}

	// return empty json
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{}")
}

func (handler *RESTInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	ReadyEntityCh = readyEntityCh
	sport := fmt.Sprintf(":%d", 10080) // FIXME
	apiRoot := "/api/v2"
	log.Debugf("REST API root=%s%s", sport, apiRoot)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", RootOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/events/{entity_id}/{event_uuid}"), EventsOnPost).Methods("POST")
	router.HandleFunc(path.Join(apiRoot, "/actions/{entity_id}"), ActionsOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/actions/{entity_id}/{action_uuid}"), ActionsOnDelete).Methods("DELETE")

	err := http.ListenAndServe(sport, router)
	if err != nil {
		panic(log.Critical(err))
	}
}

func NewRESTInspectorHanlder(config *Config) *RESTInspectorHandler {
	return &RESTInspectorHandler{}
}

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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
	"github.com/gorilla/mux"
	. "github.com/ahmetalpbalkan/go-linq"
	. "../../equtils"
)

type RESTInspectorHandler struct {
}


type Process struct {
	ActionsLock       sync.RWMutex
	Actions           []*Action
	HttpActionReadyCh chan bool
	Entity            TransitionEntity
}

//TODO: move global vars to RESTInspectorHandler struct
var (
	Procs = map[string]*Process{}
	ReadyEntityCh chan *TransitionEntity
)

func isProcessRegistered(processId string) bool {
	_, alreadyRegistered := Procs[processId]
	return alreadyRegistered
}

func registerProcess(processId string) error {
	if isProcessRegistered(processId) {
		return fmt.Errorf("process %s has been already registered", processId)
	}
	process := Process{
		ActionsLock: sync.RWMutex{},
		Actions: make([]*Action, 0),
		HttpActionReadyCh: make(chan bool),
		Entity : TransitionEntity{
			Id : processId,
			Conn: nil,
			ActionFromMain: make(chan *Action),
			EventToMain: make(chan *Event),
		},
	}
	Procs[processId] = &process
	Log("Registered process: %s", processId)
	go func() {
		for {
			// TODO: how to shutdown this goroutine
			act := <-process.Entity.ActionFromMain
			Log("(From Main)Received action: %s", act)
			process.ActionsLock.Lock()
			process.Actions = append(process.Actions, act)
			process.ActionsLock.Unlock()
			process.HttpActionReadyCh <- true
		}
	}()
	return nil
}


// @app.route('/', methods=['GET'])
func RootOnGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello Earthquake! -- RESTInspectorHandler(pid=%d)\n", os.Getpid())
}

func makeEventStruct(processId string, reader io.Reader) (ev Event, err error) {
	var m map[string] interface{}
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&m)
	if err != nil {
		return
	}
	ev, err = EventFromJSONMap(m, time.Now(), processId)
	return
}

// @app.route(api_root + '/events/<process_id>/<event_uuid>', methods=['POST'])
func EventsOnPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processId := vars["process_id"]
	eventUuid := vars["event_uuid"]
	Log("EventsOnPost: processId=%s, eventUuid=%s", processId, eventUuid)
	if ! isProcessRegistered(processId) {
		if err := registerProcess(processId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// parse event
	e, err := makeEventStruct(processId, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// send event to entity
	entity := &Procs[processId].Entity
	go func() {
		Log("*EventsOnPost: sending %s to %s", e, processId)
		entity.EventToMain <- &e
		Log("*EventsOnPost: sent %s to %s", e, processId)

	}()

	go func() {
		Log("*EventsOnPost: notifying %s to main", processId)
		ReadyEntityCh <- entity
		Log("*EventsOnPost: notified %s to main", processId)
	}()

	// return empty json
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{}")
}

func (this *Process) WaitForAction() (act *Action) {
	for {
		this.ActionsLock.RLock()
		l := len(this.Actions)
		this.ActionsLock.RUnlock()
		if l > 0 { break; }
		Log("Waiting for action (processId=%s)", this.Entity.Id)
		<-this.HttpActionReadyCh
	}
	this.ActionsLock.RLock()
	act = this.Actions[0]
	this.ActionsLock.RUnlock()
	if act.ActionType != "_JSON" {
		panic(fmt.Errorf("Invalid action type %s", act.ActionType))
	}
	return
}

// @app.route(api_root + '/actions/<process_id>', methods=['GET'])
func ActionsOnGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processId := vars["process_id"]
	Log("ActionsOnGet: processId=%s", processId)
	if ! isProcessRegistered(processId) {
		if err := registerProcess(processId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// TODO: refactor this logic
	ch := make(chan *Action)
	go func() {
		ch <- Procs[processId].WaitForAction()
	}()
	act := <-ch

	Log("Action %s ready for processId=%s", act, processId)
	// NOTE: an inspector has responsibility to DELETE the action,
	//       because HTTP DELETE must be idempotent (RFC 7231)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(act.ToJSONMap()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// @app.route(api_root + '/actions/<process_id>/<action_uuid>', methods=['DELETE'])
func ActionsOnDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processId := vars["process_id"]
	actionUuid := vars["action_uuid"]
	Log("ActionsOnDelete: processId=%s, actionUuid=%s", processId, actionUuid)
	if ! isProcessRegistered(processId) {
		err := fmt.Errorf("Unknown process %s", processId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proc := Procs[processId]
	proc.ActionsLock.Lock()
	defer proc.ActionsLock.Unlock()

	// DELETE the action
	// NOTE: newActions == Actions is expected, as DELETE must be idempotent (RFC 7231)
	newActions, err := From(proc.Actions).Where(func(s T) (bool, error){
		return s.(*Action).ToJSONMap()["uuid"].(string) != actionUuid, nil
	}).Results()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	deletedActions := len(proc.Actions) - len(newActions)
	Log("Deleted %d actions", deletedActions)
	if ! ( deletedActions == 0 || deletedActions == 1 ) {
		panic(fmt.Errorf("this should not happen. deletedActions=%d", deletedActions))
	}

	// this fails:
	//  proc.Actions = newActions.([]*Action)
	//  --> invalid type assertion: newActions.([]*equtils.Action) (non-interface type []linq.T on left)
	proc.Actions = make([]*Action, len(newActions))
	for i, _ := range newActions {
		proc.Actions[i] = newActions[i].(*Action)
	}

	// return empty json
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{}")
}

func (handler *RESTInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	Log("***** RESTInspectorHandler: UNDER WORK-IN-PROGRESS, YOU SHOULD NOT USE THIS (please try v0.1 pyearthquake instead) *****")
	ReadyEntityCh = readyEntityCh
	sport := fmt.Sprintf(":%d", 10080) // FIXME
	apiRoot := "/api/v2"
	Log("REST API root=%s%s", sport, apiRoot)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", RootOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/events/{process_id}/{event_uuid}"), EventsOnPost).Methods("POST")
	router.HandleFunc(path.Join(apiRoot, "/actions/{process_id}"), ActionsOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/actions/{process_id}/{action_uuid}"), ActionsOnDelete).Methods("DELETE")

	err := http.ListenAndServe(sport, router)
	if err != nil {
		panic(err)
	}
}

func NewRESTInspectorHanlder(config *Config) *RESTInspectorHandler {
	return &RESTInspectorHandler{}
}

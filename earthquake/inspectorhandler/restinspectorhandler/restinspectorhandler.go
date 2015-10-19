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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
	. "github.com/osrg/earthquake/earthquake/entity"
	. "github.com/osrg/earthquake/earthquake/inspectorhandler/restinspectorhandler/queue"
	. "github.com/osrg/earthquake/earthquake/signal"
	. "github.com/osrg/earthquake/earthquake/util/config"
	restutil "github.com/osrg/earthquake/earthquake/util/rest"
	"runtime"
)

var (
	mainReadyEntityCh chan *TransitionEntity
)

func newEventFromHttpRequest(r *http.Request) (Event, error) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	signal, err := NewSignalFromJSONString(string(bytes), time.Now())
	if err != nil {
		return nil, err
	}
	return signal.(Event), nil
}

func entityFromHttpRequest(r *http.Request) (*TransitionEntity, *ActionQueue, error) {
	var err error
	vars := mux.Vars(r)
	entityID := vars["entity_id"]
	entity := GetTransitionEntity(entityID)
	if entity == nil {
		entity, _, err = registerNewEntity(entityID)
		if err != nil {
			return nil, nil, err
		}
	}
	queue := GetQueue(entityID)
	if entity == nil || queue == nil {
		return nil, nil, fmt.Errorf("unexpected nil for %s", entityID)
	}
	return entity, queue, nil
}

// @app.route('/', methods=['GET'])
func rootOnGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello Earthquake! -- RESTInspectorHandler(pid=%d)\n", os.Getpid())
	w.(http.Flusher).Flush()
}

// @app.route(api_root + '/events/<entity_id>/<event_uuid>', methods=['POST'])
func eventsOnPost(w http.ResponseWriter, r *http.Request) {
	// get entity structure
	entity, _, err := entityFromHttpRequest(r)
	if err != nil {
		restutil.WriteError(w, err)
		return
	}
	// instantiate event
	event, err := newEventFromHttpRequest(r)
	if err != nil {
		restutil.WriteError(w, err)
		return
	}
	eventUUID := mux.Vars(r)["event_uuid"]
	if eventUUID != event.ID() {
		err := fmt.Errorf("uuid mismatch: %s vs %s", eventUUID, event.ID())
		restutil.WriteError(w, err)
		return
	}
	// send event to orchestrator main
	go sendEventToMain(entity, event)
	// return empty json
	if err = restutil.WriteJSON(w, map[string]interface{}{}); err != nil {
		restutil.WriteError(w, err)
	}
}

// @app.route(api_root + '/actions/<entity_id>', methods=['GET'])
//
// NOTE: an inspector has responsibility to DELETE the action, because GET must be idempotent (RFC 7231)
//
// NOTE: there should not be any concurrent GETs for an entity_id
func actionsOnGet(w http.ResponseWriter, r *http.Request) {
	var err error
	// get entity structure
	_, queue, err := entityFromHttpRequest(r)
	if err != nil {
		restutil.WriteError(w, err)
		return
	}

	// NOTE: if NumGoroutine() increases rapidly, something is going wrong.
	log.Debugf("runtime.NumGoroutine()=%d", runtime.NumGoroutine())

	// get action (this can take a while, depending on the exploration policy)
	action := queue.Peek()
	if action == nil {
		restutil.WriteError(w, fmt.Errorf("could not get action"))
		return
	}
	if err = restutil.WriteJSON(w, action.JSONMap()); err != nil {
		restutil.WriteError(w, err)
	}
}

// @app.route(api_root + '/actions/<entity_id>/<action_uuid>', methods=['DELETE'])
//
// NOTE: DELETE must be idempotent (RFC 7231)
func actionsOnDelete(w http.ResponseWriter, r *http.Request) {
	var err error
	// get entity structure
	_, queue, err := entityFromHttpRequest(r)
	if err != nil {
		restutil.WriteError(w, err)
		return
	}
	// delete the action
	actionUUID := mux.Vars(r)["action_uuid"]
	queue.Delete(actionUUID)
	// return empty json
	if err = restutil.WriteJSON(w, map[string]interface{}{}); err != nil {
		restutil.WriteError(w, err)
	}
}

type RESTInspectorHandler struct{}

func (handler *RESTInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	mainReadyEntityCh = readyEntityCh
	sport := fmt.Sprintf(":%d", restutil.DefaultPort) // FIXME
	apiRoot := restutil.APIRoot
	log.Debugf("REST API root=%s%s", sport, apiRoot)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", rootOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/events/{entity_id}/{event_uuid}"), eventsOnPost).Methods("POST")
	router.HandleFunc(path.Join(apiRoot, "/actions/{entity_id}"), actionsOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/actions/{entity_id}/{action_uuid}"), actionsOnDelete).Methods("DELETE")

	err := http.ListenAndServe(sport, router)
	if err != nil {
		panic(log.Critical(err))
	}
}

func NewRESTInspectorHanlder(config *Config) *RESTInspectorHandler {
	return &RESTInspectorHandler{}
}

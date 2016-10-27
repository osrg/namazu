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

package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"net"

	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
	. "github.com/osrg/namazu/nmz/endpoint/rest/queue"
	. "github.com/osrg/namazu/nmz/signal"
	restutil "github.com/osrg/namazu/nmz/util/rest"
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

func queueFromHttpRequest(r *http.Request) (*ActionQueue, error) {
	var err error
	vars := mux.Vars(r)
	entityID := vars["entity_id"]
	queue := GetQueue(entityID)
	if queue == nil {
		// Note that another routine can register the entity
		queue, err = RegisterNewQueue(entityID)
		// "already registered" err is not an issue here
		if err != nil && queue == nil {
			return nil, err
		}
	}
	return queue, nil
}

// @app.route('/', methods=['GET'])
func rootOnGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello Namazu! -- RESTInspectorHandler(pid=%d)\n", os.Getpid())
	w.(http.Flusher).Flush()
}

// @app.route(api_root + '/events/<entity_id>/<event_uuid>', methods=['POST'])
func eventsOnPost(w http.ResponseWriter, r *http.Request) {
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
	// register entity if it is not registered yet.
	// FIXME: rename the function
	_, err = queueFromHttpRequest(r)
	if err != nil {
		restutil.WriteError(w, err)
		return
	}

	// send event to orchestrator main
	go func() {
		orchestratorEventCh <- event
	}()
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
	// get queue
	queue, err := queueFromHttpRequest(r)
	if err != nil {
		restutil.WriteError(w, err)
		return
	}

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
	queue, err := queueFromHttpRequest(r)
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

const (
	opDisableOrchestration = "disableOrchestration"
	opEnableOrchestration  = "enableOrchestration"
)

func controlEnableOrchestration(w http.ResponseWriter, r *http.Request) {
	log.Infof("enabling orchestration")
	orchestratorControlCh <- Control{Op: ControlEnableOrchestration}
}

func controlDisableOrchestration(w http.ResponseWriter, r *http.Request) {
	log.Infof("disabling orchestration")
	orchestratorControlCh <- Control{Op: ControlDisableOrchestration}
}

func actionPropagatorRoutine() {
	for {
		action := <-orchestratorActionCh
		queue := GetQueue(action.EntityID())
		if queue == nil {
			log.Errorf("Ignored action for unknown entity %s."+
				"Orchestrator sent an action before registration is done?", action.EntityID())
			log.Errorf("Action: %#v", action)
			continue
		}
		queue.Put(action)
	}
}

type RESTEndpoint struct {
}

var (
	// this is global so that eventsOnPost() can access this
	orchestratorEventCh = make(chan Event)
	// set by Start()
	orchestratorActionCh chan Action
	// set by Start(). Useful if config port is zero.
	ActualPort int

	// used for controlling orchestrator behavior e.g. temporarily disable orchestrator
	orchestratorControlCh = make(chan Control)
)

func newRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", rootOnGet).Methods("GET")
	router.HandleFunc(path.Join(restutil.APIRoot, "/events/{entity_id}/{event_uuid}"), eventsOnPost).Methods("POST")
	router.HandleFunc(path.Join(restutil.APIRoot, "/actions/{entity_id}"), actionsOnGet).Methods("GET")
	router.HandleFunc(path.Join(restutil.APIRoot, "/actions/{entity_id}/{action_uuid}"), actionsOnDelete).Methods("DELETE")

	router.HandleFunc(path.Join(restutil.APIRoot, "/control"), controlEnableOrchestration).Queries("op", "enableOrchestration").Methods("POST")
	router.HandleFunc(path.Join(restutil.APIRoot, "/control"), controlDisableOrchestration).Queries("op", "disableOrchestration").Methods("POST")

	return router
}

// NOTE: no shutdown at the moment due to the net/http implementation issue
func (ep *RESTEndpoint) Start(port int, actionCh chan Action) (chan Event, chan Control) {
	orchestratorActionCh = actionCh
	router := newRouter()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(log.Critical(err))
	}
	ActualPort = listener.Addr().(*net.TCPAddr).Port
	if port == 0 {
		log.Infof("Automatically assigned port %d instead of 0", ActualPort)
	}
	go func() {
		err := http.Serve(listener, router)
		if err != nil {
			panic(log.Critical(err))
		}
	}()
	go actionPropagatorRoutine()
	return orchestratorEventCh, orchestratorControlCh
}

var SingletonRESTEndpoint = RESTEndpoint{}

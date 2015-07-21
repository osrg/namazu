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
	//	"encoding/json"
	"fmt"
	"os"
	"path"
	"net/http"
	//	"time"
	"github.com/gorilla/mux"
	. "../../equtils"
)

type RESTInspectorHandler struct {
}


type Process struct{
	Actions []Action
	HttpActionReadyCh chan bool
}

//TODO: move global vars to RESTInspectorHandler struct
var (
	Processes = map[string] *Process{}
)

func registerProcess(processId string) error {
	if _, alreadyRegistered := Processes[processId]; alreadyRegistered {
		return fmt.Errorf("process %s has been already registered", processId)
	}
	process := Process {
		Actions: make([]Action, 0),
		HttpActionReadyCh: make(chan bool),
	}
	Processes[processId] = &process
	Log("Registered process: %s", processId)
	return nil
}


// @app.route('/', methods=['GET'])
func RootOnGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Earthquake! -- RESTInspectorHandler(pid=%d)\n", os.Getpid())
}

// @app.route(api_root + '/events/<process_id>/<event_uuid>', methods=['POST'])
func EventsOnPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processId := vars["process_id"]
	eventUuid := vars["event_uuid"]
	Log("EventsOnPost: processId=%s, eventUuid=%s", processId, eventUuid)
	// return empty json
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{}")
}

// @app.route(api_root + '/actions/<process_id>', methods=['GET'])
func ActionsOnGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processId := vars["process_id"]
	Log("ActionsOnGet: processId=%s", processId)
	if _, alreadyRegistered := Processes[processId]; ! alreadyRegistered {
		err := registerProcess(processId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	Log("Waiting for action (processId=%s)", processId)
	<- Processes[processId].HttpActionReadyCh
	Log("Action ready for processId=%s", processId)
	//WIP
}


func (handler *RESTInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	Log("***** RESTInspectorHandler: UNDER WORK-IN-PROGRESS, YOU SHOULD NOT USE THIS (please try v0.1 pyearthquake instead) *****")
	sport := fmt.Sprintf(":%d", 10080) // FIXME
	apiRoot := "/api/v2"
	Log("REST API root=%s%s", sport, apiRoot)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", RootOnGet).Methods("GET")
	router.HandleFunc(path.Join(apiRoot, "/events/{process_id}/{event_uuid}"), EventsOnPost).Methods("POST")
	router.HandleFunc(path.Join(apiRoot, "/actions/{process_id}"), ActionsOnGet).Methods("GET")

	err := http.ListenAndServe(sport, router)
	if err != nil {
		panic(err)
	}
}

func NewRESTInspectorHanlder() *RESTInspectorHandler {
	return &RESTInspectorHandler{}
}

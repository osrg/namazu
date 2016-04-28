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

package signal

import (
	"github.com/satori/go.uuid"
)

// implements Action, OrchestratorSideAction
type NopAction struct {
	BasicAction
}

// entityID: can be empty string ("")
//
// event: can be nil
func NewNopAction(entityID string, event Event) (Action, error) {
	action := &NopAction{}
	action.InitSignal()
	action.SetID(uuid.NewV4().String())
	action.SetEntityID(entityID)
	action.SetType("action")
	action.SetClass("NopAction")
	action.CauseEvent = event
	return action, nil
}

// implements OrchestratorSideAction
func (this *NopAction) OrchestratorSideOnly() bool {
	return true
}

// implements OrchestratorSideAction
func (this *NopAction) ExecuteOnOrchestrator() error {
	return nil
}

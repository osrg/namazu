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
	"fmt"

	"github.com/satori/go.uuid"
)

// implements Action, PBAction
type EventAcceptanceAction struct {
	BasicPBAction
}

func NewEventAcceptanceAction(event Event) (Action, error) {
	action := &EventAcceptanceAction{}
	action.InitSignal()
	if !event.Deferred() {
		return action,
			fmt.Errorf("cannot instantiate EventAcceptAction for a non-deferred event %#v", event)
	}
	action.SetID(uuid.NewV4().String())
	action.SetEntityID(event.EntityID())
	action.SetType("action")
	action.SetClass("EventAcceptanceAction")
	action.Set("event_uuid", event.ID())
	action.CauseEvent = event
	return action, nil
}

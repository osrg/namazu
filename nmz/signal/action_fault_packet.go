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

// implements Action
type PacketFaultAction struct {
	BasicAction
}

func NewPacketFaultAction(event Event) (Action, error) {
	action := &PacketFaultAction{}
	action.InitSignal()
	if !event.Deferred() {
		return nil, fmt.Errorf("cannot instantiate PacketFaultAction for a non-deferred event %#v", event)
	}
	_, isPacketEvent := event.(*PacketEvent)
	if !isPacketEvent {
		return nil, fmt.Errorf("event %s is not PacketEvent", event)
	}
	action.SetID(uuid.NewV4().String())
	action.SetEntityID(event.EntityID())
	action.SetType("action")
	action.SetClass("PacketFaultAction")
	action.Set("event_uuid", event.ID())
	action.CauseEvent = event
	return action, nil
}

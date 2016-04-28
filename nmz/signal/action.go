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
	"time"

	pb "github.com/osrg/namazu/nmz/util/pb"
)

type BasicAction struct {
	BasicSignal
	// capitalized for encoding/gob
	Triggered  time.Time
	CauseEvent Event
}

// implements Action
func (this *BasicAction) Equals(o Action) bool {
	return this.EqualsSignal(o)
}

// implements Action
func (this *BasicAction) TriggeredTime() time.Time {
	return this.Triggered
}

// implements Action
func (this *BasicAction) SetTriggeredTime(triggeredTime time.Time) {
	this.Triggered = triggeredTime
}

// implements Action
//
// if only event_uuid is known, return a dummy empty event (NopEvent).
// this case happens when called from inspector
func (this *BasicAction) Event() Event {
	if this.CauseEvent != nil {
		return this.CauseEvent
	}
	eventID, ok := this.Get("event_uuid").(string)
	if !ok {
		return nil
	}
	event := NopEvent{}
	event.InitSignal()
	event.SetID(eventID)
	event.SetEntityID(this.EntityID())
	event.SetType("event")
	event.SetClass("NopEvent")
	return &event
}

// for ProtocolBuffers actions
//
// implements Action, PBAction
type BasicPBAction struct {
	BasicAction
}

// implements PBAction
func (this *BasicPBAction) PBResponseMessage() *pb.InspectorMsgRsp {
	ev := this.CauseEvent
	if ev == nil {
		return nil
	}
	pbEv, ok := ev.(*basicPBevent)
	if ok {
		return &pbEv.pbRes
	}
	return nil
}

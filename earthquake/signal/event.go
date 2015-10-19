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
	log "github.com/cihub/seelog"
	pb "github.com/osrg/earthquake/earthquake/util/pb"
)

type BasicEvent struct {
	BasicSignal
}

// implements Event
func (this *BasicEvent) Equals(o Event) bool {
	return this.Equals(o)
}

// implements Event
func (this *BasicEvent) Deferred() bool {
	deferred, ok := this.Get("deferred").(bool)
	if !ok {
		return false
	}
	return deferred
}

func (this *BasicEvent) SetDeferred(deferred bool) {
	// use BasicSignal.Set() so as to serialize in JSON
	this.Set("deferred", deferred)
}

// implements Event
func (this *BasicEvent) DefaultAction() (Action, error) {
	if this.Deferred() {
		return NewEventAcceptanceAction(this)
	}
	return NewNopAction(this.EntityID(), this)
}

// implements Event
func (this *BasicEvent) DefaultFaultAction() (Action, error) {
	return nil, nil
}

// for ProtocolBuffers events
//
// implements Signal, Event, PBEvent
type basicPBevent struct {
	BasicEvent
	pbReq pb.InspectorMsgReq
	pbRes pb.InspectorMsgRsp
}

// implements PBEvent
func (this *basicPBevent) PBRequestMessage() *pb.InspectorMsgReq {
	return &this.pbReq
}

// implements Event
// NOTE: this method should also put pb{Req,Rsp} to action.Event().
func (this *basicPBevent) DefaultAction() (Action, error) {
	if !this.Deferred() {
		panic(log.Criticalf("PBEvent is expected be deferred, %s", this))
	}
	return NewEventAcceptanceAction(this)
}

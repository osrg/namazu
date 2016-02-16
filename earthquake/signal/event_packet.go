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

import "github.com/satori/go.uuid"

// implements Event
type PacketEvent struct {
	BasicEvent
}

func NewPacketEvent(entityID, srcEntityID, dstEntityID string, m map[string]interface{}) (Event, error) {
	event := &PacketEvent{}
	event.InitSignal()
	event.SetID(uuid.NewV4().String())
	event.SetEntityID(entityID)
	event.SetType("event")
	event.SetClass("PacketEvent")
	event.SetDeferred(true)
	opt := map[string]interface{}{
		"src_entity": srcEntityID,
		"dst_entity": dstEntityID,
	}
	for k, v := range m {
		opt[k] = v
	}
	event.SetOption(opt)
	return event, nil
}

// implements Event
func (this *PacketEvent) DefaultFaultAction() (Action, error) {
	return NewPacketFaultAction(this)
}

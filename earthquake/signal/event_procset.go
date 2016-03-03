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

// Note: DefaultAction() is NopAction, not ProcSetSchedAction
// implements Event
type ProcSetEvent struct {
	BasicEvent
}

func NewProcSetEvent(entityID string, procs []string, m map[string]interface{}) (Event, error) {
	event := &ProcSetEvent{}
	event.InitSignal()
	event.SetID(uuid.NewV4().String())
	event.SetEntityID(entityID)
	event.SetType("event")
	event.SetClass("ProcSetEvent")
	event.SetDeferred(false)
	opt := map[string]interface{}{
		"procs": procs,
	}
	for k, v := range m {
		opt[k] = v
	}
	event.SetOption(opt)
	return event, nil
}

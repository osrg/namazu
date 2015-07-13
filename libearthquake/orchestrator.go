// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

package main

import (
	. "./equtils"
	. "./explorepolicy"
	"./inspectorhandler"
)

func orchestrate(endCh chan interface{}, policy ExplorePolicy, newTraceCh chan *SingleTrace) {
	readyEntityCh := make(chan *TransitionEntity)

	inspectorhandler.StartAllInspectorHandler(readyEntityCh)

	eventSeq := make([]Event, 0)

	nextEventChan := policy.GetNextEventChan()
	ev2entity := make(map[*Event]*TransitionEntity)

	Log("start execution loop body")

	running := true
	for {
		select {
		case readyEntity := <-readyEntityCh:
			if !running {
				// run script ended, just ignore events
				readyEntity.GotoNext <- true
				continue
			}

			Log("ready process %v", readyEntity)
			event := <-readyEntity.EventToMain
			Log("recieved message from %v", readyEntity)

			ev2entity[event] = readyEntity
			policy.QueueNextEvent(readyEntity.Id, event)

		case nextEvent := <-nextEventChan:
			readyEntity := ev2entity[nextEvent]
			delete(ev2entity, nextEvent)

			eventSeq = append(eventSeq, *nextEvent)
			readyEntity.GotoNext <- true

		case <-endCh:
			Log("main loop end")
			running = false

			newTrace := &SingleTrace{
				eventSeq,
			}
			newTraceCh <- newTrace
		}
	}

	Log("end execution loop body")
}

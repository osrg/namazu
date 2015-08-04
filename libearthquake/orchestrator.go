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

func orchestrate(endCh chan interface{}, policy ExplorePolicy, newTraceCh chan *SingleTrace, config *Config) {
	readyEntityCh := make(chan *TransitionEntity)

	inspectorhandler.StartAllInspectorHandler(readyEntityCh, config)

	actionSeq := make([]Action, 0)

	nextActionChan := policy.GetNextActionChan()

	knownEntities := make(map[string]*TransitionEntity)

	Log("start execution loop body")

	running := true

	handleAction := func(nextAction *Action) {
		if nextAction.OrchestratorLocal {
			Log("main loop received action (type=\"%s\") from the policy. " +
			"NOT passing to the inspector handler", nextAction.ActionType)
		} else {
			Log("main loop received action (type=\"%s\") from the policy. " +
			"passing to the inspector handler", nextAction.ActionType)
		}
		// find corresponding entity
		readyEntity := knownEntities[nextAction.EntityId]

		// make sequence for tracing
		actionSeq = append(actionSeq, *nextAction)

		// pass to the inspector handler.
		// inspector handler should verify action.
		if ! nextAction.OrchestratorLocal {
			readyEntity.ActionFromMain <- nextAction
		}
	}

	for {
		select {
		case readyEntity := <-readyEntityCh:
			Log("ready entity %v", readyEntity)
			event := <-readyEntity.EventToMain
			Log("recieved message from %v", readyEntity)

			acceptEventImmediately := func(event *Event) {
				act, err := event.MakeAcceptAction()
				if err != nil {
					panic(err)
				}
				handleAction(act)
			}

			if running {
				knownEntities[readyEntity.Id] = readyEntity
				if event.Deferred {
					policy.QueueNextEvent(readyEntity.Id, event)
				} else {
					// not deferred events: e.g. syslog
					acceptEventImmediately(event)
				}
			} else  {
				// run script ended, accept event immediately without passing to the policy
				acceptEventImmediately(event)
			}

		case nextAction := <-nextActionChan:
			handleAction(nextAction)

		case <-endCh:
			Log("main loop end")
			running = false

			newTrace := &SingleTrace{
				actionSeq,
			}
			newTraceCh <- newTrace
		}
	}

	Log("end execution loop body")
}

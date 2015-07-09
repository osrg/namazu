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
	"time"

	. "./searchpolicy"
	"./inspectorhandler"
)

func orchestrate(endCh chan interface{}, policy SearchPolicy, newTraceCh chan *SingleTrace) {
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
			req := <-readyEntity.ReqToMain
			eventReq := req.Event
			Log("recieved message from %v", readyEntity)

			if *eventReq.Type == InspectorMsgReq_Event_EXIT {
				Log("process %v is exiting", readyEntity)
				continue
			}

			e := &Event{
				ArrivedTime: time.Now(),
				ProcId:      readyEntity.Id,

				EventType:  "FuncCall",
				EventParam: *eventReq.FuncCall.Name,
			}

			if *req.HasJavaSpecificFields == 1 {
				ejs := Event_JavaSpecific{
					ThreadName: *req.JavaSpecificFields.ThreadName,
				}

				for _, stackTraceElement := range req.JavaSpecificFields.StackTraceElements {
					element := Event_JavaSpecific_StackTraceElement{
						LineNumber: int(*stackTraceElement.LineNumber),
						ClassName:  *stackTraceElement.ClassName,
						MethodName: *stackTraceElement.MethodName,
						FileName:   *stackTraceElement.FileName,
					}

					ejs.StackTraceElements = append(ejs.StackTraceElements, element)
				}
				ejs.NrStackTraceElements = int(*req.JavaSpecificFields.NrStackTraceElements)

				for _, param := range req.JavaSpecificFields.Params {
					param := Event_JavaSpecific_Param{
						Name:  *param.Name,
						Value: *param.Value,
					}

					ejs.Params = append(ejs.Params, param)
				}

				ejs.NrParams = int(*req.JavaSpecificFields.NrParams)

				e.JavaSpecific = &ejs
			}

			ev2entity[e] = readyEntity
			policy.QueueNextEvent(readyEntity.Id, e)

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

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
	"fmt"
	"io"
	"net"
	"os"
	"time"

	. "./searchpolicy"
)

type transitionEntity struct {
	id   string
	conn net.Conn

	reqToMain chan *InspectorMsgReq
	gotoNext  chan interface{}

	eventReqRecv chan *InspectorMsgReq
	eventRspSend chan *InspectorMsgRsp
}

func handleProcessNoInitiation(proc *transitionEntity, readyEntityCh chan *transitionEntity) {
	go func(p *transitionEntity) {
		for {
			req := &InspectorMsgReq{}

			rerr := RecvMsg(p.conn, req)
			if rerr != nil {
				if rerr == io.EOF {
					Log("received EOF from transition entity :%s", p.id)
					return
				} else {
					Log("failed to recieve request (transition entity: %s): %s", p.id, rerr)
					return // TODO: error handling
				}
			}

			Log("received message from transition entity :%s", p.id)
			p.eventReqRecv <- req
		}
	}(proc)

	go func(p *transitionEntity) {
		for {
			rsp := <-p.eventRspSend
			serr := SendMsg(p.conn, rsp)
			if serr != nil {
				Log("failed to send response (transition entity: %s): %s", p.id, serr)
				return // TODO: error handling
			}
			if *rsp.Res == InspectorMsgRsp_END {
				Log("send routine end (transition entity :%s)", p.id)
				return
			}
		}
	}(proc)

	recvCh := make(chan bool)

	req := (*InspectorMsgReq)(nil)

	go func() {
		for {
			req = <-proc.eventReqRecv
			Log("received event from main goroutine: %s", proc.id)
			recvCh <- true
		}
	}()

	for {
		select {
		case <-recvCh:
			if *req.Type != InspectorMsgReq_EVENT {
				Log("invalid message from transition entity %s, type: %d", proc.id, *req.Type)
				os.Exit(1)
			}

			if proc.id == "uninitialized" {
				// initialize id with a member of event
				proc.id = *req.ProcessId
			}

			Log("event message received from transition entity %s", proc.id)

			go func(r *InspectorMsgReq) {
				proc.reqToMain <- r
			}(req)

			readyEntityCh <- proc
			if *req.Event.Type != InspectorMsgReq_Event_EXIT {
				<-proc.gotoNext

				result := InspectorMsgRsp_ACK
				req_msg_id := *req.MsgId
				rsp := &InspectorMsgRsp{
					Res:   &result,
					MsgId: &req_msg_id,
				}

				proc.eventRspSend <- rsp
				Log("replied to the event message from process %v", proc)
			}
		}
	}
}

func acceptNewProcess(readyEntityCh chan *transitionEntity) {
	sport := fmt.Sprintf(":%d", 10000) // FIXME
	ln, lerr := net.Listen("tcp", sport)
	if lerr != nil {
		Log("failed to listen on port %d: %s", 10000, lerr)
		os.Exit(1)
	}

	for {
		conn, aerr := ln.Accept()
		if aerr != nil {
			Log("failed to accept on %v: %s", ln, aerr)
			os.Exit(1)
		}

		Log("accepted new connection: %v", conn)

		proc := new(transitionEntity)
		proc.id = "uninitialized"
		proc.conn = conn
		proc.gotoNext = make(chan interface{})
		proc.eventReqRecv = make(chan *InspectorMsgReq)
		proc.eventRspSend = make(chan *InspectorMsgRsp)
		proc.reqToMain = make(chan *InspectorMsgReq)

		go handleProcessNoInitiation(proc, readyEntityCh)
	}
}

func singleSearchNoInitiation(workingDir string, endCh chan interface{}, policy SearchPolicy, newTraceCh chan *SingleTrace) {
	readyEntityCh := make(chan *transitionEntity)

	go acceptNewProcess(readyEntityCh)

	eventSeq := make([]Event, 0)

	nextEventChan := policy.GetNextEventChan()
	ev2entity := make(map[*Event]*transitionEntity)

	running := true
	for {
		select {
		case readyEntity := <-readyEntityCh:
			if !running {
				// run script ended, just ignore events
				readyEntity.gotoNext <- true
				continue
			}

			Log("ready process %v", readyEntity)
			req := <-readyEntity.reqToMain
			eventReq := req.Event
			Log("recieved message from %v", readyEntity)

			if *eventReq.Type == InspectorMsgReq_Event_EXIT {
				Log("process %v is exiting", readyEntity)
				continue
			}

			e := &Event{
				ArrivedTime: time.Now(),
				ProcId:      readyEntity.id,

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
			policy.QueueNextEvent(readyEntity.id, e)

		case nextEvent := <-nextEventChan:
			readyEntity := ev2entity[nextEvent]
			delete(ev2entity, nextEvent)

			eventSeq = append(eventSeq, *nextEvent)
			readyEntity.gotoNext <- true

		case <-endCh:
			Log("main loop end")
			running = false

			newTrace := &SingleTrace{
				eventSeq,
			}
			newTraceCh <- newTrace
		}
	}
}

func searchModeNoInitiation(workingDir string, policy SearchPolicy, endCh chan interface{}, newTraceCh chan *SingleTrace) {
	Log("start execution loop body")
	singleSearchNoInitiation(workingDir, endCh, policy, newTraceCh)
	Log("end execution loop body")
}

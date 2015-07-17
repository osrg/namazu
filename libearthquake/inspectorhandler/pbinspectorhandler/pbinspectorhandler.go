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

package pbinspectorhandler

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	. "../../equtils"
)

type PBInspectorHandler struct {
}

func (handler *PBInspectorHandler) handleEntity(entity *TransitionEntity, readyEntityCh chan *TransitionEntity) {
	eventReqRecv := make(chan *InspectorMsgReq)
	eventRspSend := make(chan *InspectorMsgRsp)

	go func(e *TransitionEntity) {
		for {
			req := &InspectorMsgReq{}

			rerr := RecvMsg(e.Conn, req)
			if rerr != nil {
				if rerr == io.EOF {
					Log("received EOF from transition entity :%s", e.Id)
					return
				} else {
					Log("failed to recieve request (transition entity: %s): %s", e.Id, rerr)
					return // TODO: error handling
				}
			}

			Log("received message from transition entity :%s", e.Id)
			eventReqRecv <- req
		}
	}(entity)

	go func(e *TransitionEntity) {
		for {
			rsp := <-eventRspSend
			serr := SendMsg(e.Conn, rsp)
			if serr != nil {
				Log("failed to send response (transition entity: %s): %s", e.Id, serr)
				return // TODO: error handling
			}
			if *rsp.Res == InspectorMsgRsp_END {
				Log("send routine end (transition entity :%s)", e.Id)
				return
			}
		}
	}(entity)

	for {
		select {
		case req := <-eventReqRecv:
			if *req.Type != InspectorMsgReq_EVENT {
				Log("invalid message from transition entity %s, type: %d", entity.Id, *req.Type)
				os.Exit(1)
			}

			if *req.Event.Type == InspectorMsgReq_Event_EXIT {
				Log("process %v is exiting", entity)
				continue
			}

			if entity.Id == "uninitialized" {
				// initialize id with a member of event
				entity.Id = *req.ProcessId
			}

			Log("event message received from transition entity %s", entity.Id)

			evType := ""
			evParam := ""
			if *req.Event.Type == InspectorMsgReq_Event_FUNC_CALL {
				evType = "FuncCall"
				evParam = *req.Event.FuncCall.Name
			} else if *req.Event.Type == InspectorMsgReq_Event_FUNC_RETURN {
				evType = "FuncReturn"
				evParam = *req.Event.FuncReturn.Name
			} else {
				Panic("invalid type of event: %d", *req.Event.Type)
			}

			e := &Event{
				ArrivedTime: time.Now(),
				ProcId:      entity.Id,

				EventType:  evType,
				EventParam: evParam,
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

			go func(ev *Event) {
				entity.EventToMain <- ev
			}(e)

			readyEntityCh <- entity

			if *req.Event.Type != InspectorMsgReq_Event_EXIT {
				act := <-entity.ActionFromMain
				Log("execute action (type=\"%s\")", act.ActionType)
				// FIXME: wrap action type string
				if (act.ActionType == "Accept") {
					result := InspectorMsgRsp_ACK
					req_msg_id := *req.MsgId
					rsp := &InspectorMsgRsp{
						Res:   &result,
						MsgId: &req_msg_id,
					}
					eventRspSend <- rsp
					Log("accepted the event message from process %v", entity)
				} else {
					Panic("unsupported action %s", act.ActionType)
				}
			}
		}
	}
}

func (handler *PBInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
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

		entity := new(TransitionEntity)
		entity.Id = "uninitialized"
		entity.Conn = conn
		entity.ActionFromMain = make(chan *Action)
		entity.EventToMain = make(chan *Event)

		go handler.handleEntity(entity, readyEntityCh)
	}
}

func NewPBInspectorHanlder() *PBInspectorHandler {
	return &PBInspectorHandler{}
}

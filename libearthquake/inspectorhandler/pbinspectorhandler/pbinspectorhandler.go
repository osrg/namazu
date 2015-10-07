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
	. "../../equtils"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/satori/go.uuid"
	"io"
	"net"
	"os"
	"time"
)

type PBInspectorHandler struct {
}

func recvPBMsgViaChan(entity *TransitionEntity, eventReqRecv chan *InspectorMsgReq) {
	for {
		req := &InspectorMsgReq{}

		rerr := RecvMsg(entity.Conn, req)
		if rerr != nil {
			if rerr == io.EOF {
				log.Debugf("received EOF from transition entity :%s", entity.Id)
				return
			} else {
				log.Debugf("failed to recieve request (transition entity: %s): %s", entity.Id, rerr)
				return // TODO: error handling
			}
		}

		log.Debugf("received message from transition entity :%s", entity.Id)
		eventReqRecv <- req
	}
}

func sendPBMsgViaChan(entity *TransitionEntity, eventRspSend chan *InspectorMsgRsp) {
	for {
		rsp := <-eventRspSend
		serr := SendMsg(entity.Conn, rsp)
		if serr != nil {
			log.Debugf("failed to send response (transition entity: %s): %s", entity.Id, serr)
			return // TODO: error handling
		}
		if *rsp.Res == InspectorMsgRsp_END {
			log.Debugf("send routine end (transition entity :%s)", entity.Id)
			return
		}
	}
}

func (handler *PBInspectorHandler) makeEventFromPBMsg(entity *TransitionEntity, req *InspectorMsgReq) *Event {

	evType := ""
	evParam := NewEAParam()
	evDeferred := false
	if *req.Event.Type == InspectorMsgReq_Event_FUNC_CALL {
		evType = "FuncCall"
		evParam["name"] = *req.Event.FuncCall.Name
		evDeferred = true
	} else if *req.Event.Type == InspectorMsgReq_Event_FUNC_RETURN {
		evType = "FuncReturn"
		evParam["name"] = *req.Event.FuncReturn.Name
		evDeferred = true
	} else {
		panic(log.Criticalf("invalid type of event: %d", *req.Event.Type))
	}

	e := &Event{
		ArrivedTime: time.Now(),
		EntityId:    entity.Id,

		// eventId: used by MongoDB and so on. expected to compliant with RFC 4122 UUID string format
		EventId:    uuid.NewV4().String(),
		EventType:  evType,
		EventParam: evParam,
		Deferred:   evDeferred,
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
	if err := e.Validate(); err != nil {
		panic(log.Criticalf("Event.Validate() returned an error %s", err))
	}
	return e
}

func (handler *PBInspectorHandler) makePBMsgFromAction(entity *TransitionEntity, req *InspectorMsgReq, action *Action) *InspectorMsgRsp {
	if err := action.Validate(); err != nil {
		panic(log.Critical(err))
	}
	if action.ActionType == "Accept" {
		result := InspectorMsgRsp_ACK
		req_msg_id := *req.MsgId
		rsp := &InspectorMsgRsp{
			Res:   &result,
			MsgId: &req_msg_id,
		}
		return rsp
	}
	panic(log.Criticalf("unsupported action %s", action))
	return nil
}

func (handler *PBInspectorHandler) handleEntity(entity *TransitionEntity, readyEntityCh chan *TransitionEntity) {
	eventReqRecv := make(chan *InspectorMsgReq)
	eventRspSend := make(chan *InspectorMsgRsp)
	go recvPBMsgViaChan(entity, eventReqRecv)
	go sendPBMsgViaChan(entity, eventRspSend)

	for {
		select {
		case req := <-eventReqRecv:
			if *req.Type != InspectorMsgReq_EVENT {
				log.Debugf("invalid message from transition entity %s, type: %d", entity.Id, *req.Type)
				os.Exit(1)
			}

			if *req.Event.Type == InspectorMsgReq_Event_EXIT {
				log.Debugf("entity %v is exiting", entity)
				continue
			}

			if entity.Id == "uninitialized" {
				// initialize id with a member of event
				entity.Id = *req.EntityId
			}
			log.Debugf("event message received from transition entity %s", entity.Id)

			e := handler.makeEventFromPBMsg(entity, req)
			go func(ev *Event) {
				entity.EventToMain <- ev
			}(e)
			readyEntityCh <- entity

			if *req.Event.Type != InspectorMsgReq_Event_EXIT {
				act := <-entity.ActionFromMain
				log.Debugf("execute action (type=\"%s\")", act.ActionType)
				rsp := handler.makePBMsgFromAction(entity, req, act)
				eventRspSend <- rsp
				log.Debugf("accepted the event message from entity %v", entity)
			}
		} // select
	} // for
} // func

func (handler *PBInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	sport := fmt.Sprintf(":%d", 10000) // FIXME (config.GetInt("inspectorHandler.pb.port"))
	ln, lerr := net.Listen("tcp", sport)
	if lerr != nil {
		panic(log.Criticalf("failed to listen on port %d: %s", 10000, lerr))
	}

	for {
		conn, aerr := ln.Accept()
		if aerr != nil {
			panic(log.Criticalf("failed to accept on %v: %s", ln, aerr))
		}

		log.Debugf("accepted new connection: %v", conn)

		entity := new(TransitionEntity)
		entity.Id = "uninitialized"
		entity.Conn = conn
		entity.ActionFromMain = make(chan *Action)
		entity.EventToMain = make(chan *Event)

		go handler.handleEntity(entity, readyEntityCh)
	}
}

func NewPBInspectorHanlder(config *Config) *PBInspectorHandler {
	return &PBInspectorHandler{}
}

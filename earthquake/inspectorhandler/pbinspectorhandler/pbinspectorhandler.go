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
	"time"

	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/entity"
	. "github.com/osrg/earthquake/earthquake/signal"
	. "github.com/osrg/earthquake/earthquake/util/pb"
)

type PBInspectorHandler struct {
	Port int
}

func recvPBMsgViaChan(conn net.Conn, eventReqRecv chan *InspectorMsgReq) {
	for {
		req := &InspectorMsgReq{}
		rerr := RecvMsg(conn, req)
		if rerr != nil {
			if rerr == io.EOF {
				log.Debugf("received EOF from %s", conn.RemoteAddr())
				return
			} else {
				log.Errorf("failed to recieve request %s", conn.RemoteAddr())
				return // TODO: error handling
			}
		}
		log.Debugf("received message from %s", conn.RemoteAddr())
		eventReqRecv <- req
	}
}

func sendPBMsgViaChan(conn net.Conn, eventRspSend chan *InspectorMsgRsp) {
	for {
		rsp := <-eventRspSend
		err := SendMsg(conn, rsp)
		if err != nil {
			log.Errorf("failed to send response to %s: %s", conn.RemoteAddr(), err)
			return // TODO: error handling
		}
		if *rsp.Res == InspectorMsgRsp_END {
			log.Debugf("send routine end (%s)", conn.RemoteAddr())
			return
		}
	}
}

func (handler *PBInspectorHandler) handleConn(conn net.Conn, readyEntityCh chan *TransitionEntity) {
	msgReqRecvCh := make(chan *InspectorMsgReq)
	msgRspSendCh := make(chan *InspectorMsgRsp)
	go recvPBMsgViaChan(conn, msgReqRecvCh)
	go sendPBMsgViaChan(conn, msgRspSendCh)

	for {
		log.Debugf("Handler[%s]<-Inspector: receiving an event", conn.RemoteAddr())
		pbReqMsg := <-msgReqRecvCh
		entityID := *pbReqMsg.EntityId
		if *pbReqMsg.Type != InspectorMsgReq_EVENT {
			panic(log.Criticalf("Handler[%s,%s]: invalid message type: %d", conn.RemoteAddr(), entityID, *pbReqMsg.Type))
		}

		if *pbReqMsg.Event.Type == InspectorMsgReq_Event_EXIT {
			log.Debugf("Handler[%s,%s]: entity is exiting", conn.RemoteAddr(), entityID)
			continue
		}

		entity := GetTransitionEntity(entityID)
		if entity == nil {
			// orchestrator requires this registration
			entity = &TransitionEntity{
				ID:             entityID,
				ActionFromMain: make(chan Action),
				EventToMain:    make(chan Event),
			}
			err := RegisterTransitionEntity(entity)
			if err != nil {
				panic(log.Critical(err))
			}
			log.Debugf("Handler[%s,%s]: initialized the entity structure", conn.RemoteAddr(), entityID)
		}

		event, err := NewJavaFunctionEventFromPB(*pbReqMsg, time.Now())
		if err != nil {
			panic(log.Critical(err))
		}
		log.Debugf("Handler[%s,%s]<-Inspector: received an event %s", conn.RemoteAddr(), entityID, event)

		log.Debugf("Handler[%s,%s]->Main: sending an event %s", conn.RemoteAddr(), entityID, event)
		go func() {
			entity.EventToMain <- event
		}()
		readyEntityCh <- entity
		log.Debugf("Handler[%s,%s]->Main: sent an event %s", conn.RemoteAddr(), entityID, event)

		log.Debugf("Handler[%s,%s]<-Main: receiving an action", conn.RemoteAddr(), entityID)
		action := <-entity.ActionFromMain
		log.Debugf("Handler[%s,%s]<-Main: received an action %s", conn.RemoteAddr(), entityID, action)

		pbAction, pbActionOk := action.(PBAction)
		if !pbActionOk {
			panic(log.Criticalf("cannot convert %s to PBAction", action))
		}
		pbRspMsg := pbAction.PBResponseMessage()
		if pbRspMsg == nil {
			panic(log.Criticalf("pbRspMsg is nil, pbAction=%s", pbAction))
		}

		log.Debugf("Handler[%s,%s]->Inspector: sending an action %s", conn.RemoteAddr(), entityID, action)
		msgRspSendCh <- pbRspMsg
		log.Debugf("Handler[%s,%s]->Inspector: sent an action %s", conn.RemoteAddr(), entityID, action)
	} // for
} // func

func (handler *PBInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	sport := fmt.Sprintf(":%d", handler.Port)
	ln, err := net.Listen("tcp", sport)
	if err != nil {
		panic(log.Criticalf("failed to listen on port %s: %s", sport, err))
	}
	log.Debugf("PB root=%s", sport)

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(log.Criticalf("failed to accept on %v: %s", ln, err))
		}
		log.Debugf("Handler: accepted new connection from %s", conn.RemoteAddr())
		go handler.handleConn(conn, readyEntityCh)
	}
}

func NewPBInspectorHanlder(port int) *PBInspectorHandler {
	return &PBInspectorHandler{
		Port: port,
	}
}

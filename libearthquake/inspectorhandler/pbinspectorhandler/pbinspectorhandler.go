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

	. "../../equtils"
)

type PBInspectorHandler struct {
}

func (handler *PBInspectorHandler) handleEntity(entity *TransitionEntity, readyEntityCh chan *TransitionEntity) {
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
			e.EventReqRecv <- req
		}
	}(entity)

	go func(e *TransitionEntity) {
		for {
			rsp := <-e.EventRspSend
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

	recvCh := make(chan bool)

	req := (*InspectorMsgReq)(nil)

	go func() {
		for {
			req = <-entity.EventReqRecv
			Log("received event from main goroutine: %s", entity.Id)
			recvCh <- true
		}
	}()

	for {
		select {
		case <-recvCh:
			if *req.Type != InspectorMsgReq_EVENT {
				Log("invalid message from transition entity %s, type: %d", entity.Id, *req.Type)
				os.Exit(1)
			}

			if entity.Id == "uninitialized" {
				// initialize id with a member of event
				entity.Id = *req.ProcessId
			}

			Log("event message received from transition entity %s", entity.Id)

			go func(r *InspectorMsgReq) {
				entity.ReqToMain <- r
			}(req)

			readyEntityCh <- entity
			if *req.Event.Type != InspectorMsgReq_Event_EXIT {
				<-entity.GotoNext

				result := InspectorMsgRsp_ACK
				req_msg_id := *req.MsgId
				rsp := &InspectorMsgRsp{
					Res:   &result,
					MsgId: &req_msg_id,
				}

				entity.EventRspSend <- rsp
				Log("replied to the event message from process %v", entity)
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
		entity.GotoNext = make(chan interface{})
		entity.EventReqRecv = make(chan *InspectorMsgReq)
		entity.EventRspSend = make(chan *InspectorMsgRsp)
		entity.ReqToMain = make(chan *InspectorMsgReq)

		go handler.handleEntity(entity, readyEntityCh)
	}
}

func NewPBInspectorHanlder() *PBInspectorHandler {
	return &PBInspectorHandler{}
}

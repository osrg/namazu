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

package pb

import (
	"fmt"
	log "github.com/cihub/seelog"
	. "github.com/osrg/namazu/nmz/signal"
	. "github.com/osrg/namazu/nmz/util/pb"
	"io"
	"net"
	"sync"
	"time"
)

type PBEndpoint struct {
	// FIXME: entity id string cannot be shared between diferrent net.Conns (i.e. not retrieable)
	actionChs           map[string]chan Action
	actionChsMu         *sync.RWMutex
	orchestratorEventCh chan Event
	// set by Start()
	orchestratorActionCh chan Action
	// set by Start(). Useful if config port is zero.
	ActualPort int
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
				log.Errorf("failed to recieve request %s: %s", conn.RemoteAddr(), rerr)
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

func (ep *PBEndpoint) handleEntity(entityID string, msgRspSendCh chan *InspectorMsgRsp) {
	actionCh := make(chan Action)
	ep.actionChsMu.Lock()
	_, ok := ep.actionChs[entityID]
	if ok {
		ep.actionChsMu.Unlock()
		return
	}
	ep.actionChs[entityID] = actionCh
	ep.actionChsMu.Unlock()
	go func() {
		for {
			action := <-actionCh
			pbAction, pbActionOk := action.(PBAction)
			if !pbActionOk {
				panic(log.Criticalf("cannot convert %s to PBAction", action))
			}
			pbRspMsg := pbAction.PBResponseMessage()
			if pbRspMsg == nil {
				panic(log.Criticalf("pbRspMsg is nil, pbAction=%s", pbAction))
			}
			msgRspSendCh <- pbRspMsg
		}
	}()
}

func (ep *PBEndpoint) connRoutine(conn net.Conn) {
	msgReqRecvCh := make(chan *InspectorMsgReq)
	msgRspSendCh := make(chan *InspectorMsgRsp)
	go recvPBMsgViaChan(conn, msgReqRecvCh)
	go sendPBMsgViaChan(conn, msgRspSendCh)

	for {
		log.Debugf("EP[%s]: receiving", conn.RemoteAddr())
		pbReqMsg := <-msgReqRecvCh
		entityID := *pbReqMsg.EntityId
		if *pbReqMsg.Type != InspectorMsgReq_EVENT {
			panic(log.Criticalf("EP[%s,%s]: invalid message type: %d", conn.RemoteAddr(), entityID, *pbReqMsg.Type))
		}
		log.Debugf("EP[%s,%s]: received %s", conn.RemoteAddr(), entityID, pbReqMsg)
		if *pbReqMsg.Event.Type == InspectorMsgReq_Event_EXIT {
			continue
		}
		ep.handleEntity(entityID, msgRspSendCh)
		event, err := NewJavaFunctionEventFromPB(*pbReqMsg, time.Now())
		if err != nil {
			panic(log.Critical(err))
		}
		ep.orchestratorEventCh <- event
	} // for
} // func

func (ep *PBEndpoint) actionRoutine() {
	for {
		select {
		case action, ok := <-ep.orchestratorActionCh:
			if ok {
				ep.actionChsMu.RLock()
				actionCh, actionChOk := ep.actionChs[action.EntityID()]
				ep.actionChsMu.RUnlock()
				if !actionChOk {
					panic(log.Criticalf("unknown entity: %s", action.EntityID()))
				}
				actionCh <- action
			}
		}
	}
}

func (ep *PBEndpoint) Start(port int, actionCh chan Action) chan Event {
	ep.orchestratorActionCh = actionCh
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(log.Critical(err))
	}
	ep.ActualPort = listener.Addr().(*net.TCPAddr).Port
	if port == 0 {
		log.Infof("Automatically assigned port %d instead of 0", ep.ActualPort)
	}
	go ep.actionRoutine()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(log.Criticalf("failed to accept on %v: %s", listener, err))
			}
			log.Debugf("accepted new connection from %s", conn.RemoteAddr())
			go ep.connRoutine(conn)
		}
	}()
	return ep.orchestratorEventCh
}

func NewPBEndpoint() PBEndpoint {
	return PBEndpoint{
		actionChs:           make(map[string]chan Action),
		actionChsMu:         &sync.RWMutex{},
		orchestratorEventCh: make(chan Event),
	}
}

var SingletonPBEndpoint = NewPBEndpoint()

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
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/osrg/namazu/nmz/signal"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/osrg/namazu/nmz/util/mockorchestrator"
	pbutil "github.com/osrg/namazu/nmz/util/pb"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	actionCh := make(chan signal.Action)
	eventCh := SingletonPBEndpoint.Start(0, actionCh)
	mockOrc := mockorchestrator.NewMockOrchestrator(eventCh, actionCh)
	mockOrc.Start()
	defer mockOrc.Shutdown()
	os.Exit(m.Run())
}

func newPBMessage(t *testing.T, entityID string, value int) proto.Message {
	dummyFuncName := fmt.Sprintf("func-%d", value)
	pbType := pbutil.InspectorMsgReq_EVENT
	pbFuncType := pbutil.InspectorMsgReq_Event_FUNC_CALL
	pbReq := pbutil.InspectorMsgReq{
		Type:     &pbType,
		EntityId: proto.String(entityID),
		Pid:      proto.Int32(42),
		Tid:      proto.Int32(42),
		MsgId:    proto.Int32(int32(value)),
		Event: &pbutil.InspectorMsgReq_Event{
			Type: &pbFuncType,
			FuncCall: &pbutil.InspectorMsgReq_Event_FuncCall{
				Name: proto.String(dummyFuncName),
			},
		},
		HasJavaSpecificFields: proto.Int32(0),
	}
	return &pbReq
}

func dial(t *testing.T) net.Conn {
	s := fmt.Sprintf(":%d", SingletonPBEndpoint.ActualPort)
	t.Logf("Dialing to %s", s)
	conn, err := net.Dial("tcp", s)
	assert.NoError(t, err)
	return conn
}

func TestPBEndpoint_1_1(t *testing.T) {
	testPBEndpoint(t, 1, 1, false)
}

func TestPBEndpoint_1_2(t *testing.T) {
	testPBEndpoint(t, 1, 2, false)
}

func TestPBEndpoint_2_1(t *testing.T) {
	testPBEndpoint(t, 2, 1, false)
}

func TestPBEndpoint_2_2(t *testing.T) {
	testPBEndpoint(t, 2, 2, false)
}

func TestPBEndpoint_10_10(t *testing.T) {
	testPBEndpoint(t, 10, 10, false)
}

func TestPBEndpointShouldNotBlock_10_10(t *testing.T) {
	testPBEndpoint(t, 10, 10, true)
}

func testPBEndpoint(t *testing.T, n, entities int, concurrent bool) {
	conns := make([]net.Conn, entities)
	for i := 0; i < entities; i++ {
		conns[i] = dial(t)
	}

	sender := func() {
		for i := 0; i < n; i++ {
			// FIXME: entity id string cannot be shared between diferrent net.Conns (i.e. not retrieable)
			// so we append time.Now() here at the moment
			entityID := fmt.Sprintf("entity-%d-%s", i%entities, time.Now())
			req := newPBMessage(t, entityID, i)
			t.Logf("Test %d: Sending %s", i, req)
			err := pbutil.SendMsg(conns[i%entities], req)
			assert.NoError(t, err)
			t.Logf("Test %d: Sent %s", i, req)
		}
	}
	receiver := func() {
		for i := 0; i < n; i++ {
			t.Logf("Test %d: Receiving", i)
			rsp := pbutil.InspectorMsgRsp{}
			err := pbutil.RecvMsg(conns[i%entities], &rsp)
			assert.NoError(t, err)
			t.Logf("Test %d: Received %v", i, rsp)
		}
	}
	if concurrent {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			sender()
		}()
		go func() {
			defer wg.Done()
			receiver()
		}()
		wg.Wait()
	} else {
		sender()
		receiver()
	}
}

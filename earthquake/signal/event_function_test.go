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

package signal

import (
	"github.com/golang/protobuf/proto"
	"github.com/osrg/earthquake/earthquake/util/pb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewJavaFunctionEventFromPB(t *testing.T) {
	pbType := pb.InspectorMsgReq_EVENT
	pbFuncType := pb.InspectorMsgReq_Event_FUNC_CALL
	pbReq := pb.InspectorMsgReq{
		Type:     &pbType,
		EntityId: proto.String("dummy"),
		Event: &pb.InspectorMsgReq_Event{
			Type: &pbFuncType,
			FuncCall: &pb.InspectorMsgReq_Event_FuncCall{
				Name: proto.String("sayHello"),
			},
		},
	}

	event, err := NewJavaFunctionEventFromPB(pbReq, time.Now())
	assert.NoError(t, err)
	t.Logf("event: %#v", event)
	javaFunctionEvent, ok := event.(*JavaFunctionEvent)
	if !ok {
		t.Fatal("Cannot convert to JavaFunctionEvent")
	}
	assert.Equal(t, "dummy", event.EntityID())
	assert.Equal(t, "sayHello", javaFunctionEvent.FunctionName)
	t.Logf("event JSON map: %#v", event.JSONMap())

	action := testDeferredEventDefaultAction(t, event)
	testGOBAction(t, action, event)

	pbAction, pbActionOk := action.(PBAction)
	if !pbActionOk {
		t.Fatalf("non-PBAction action: %#v", action)
	}
	pbRspMsg := pbAction.PBResponseMessage()
	assert.NotNil(t, pbRspMsg, "pbRspMsg should not be nil, action=%#v", action)
	t.Logf("pbRspMsg: %s", pbRspMsg)
}

func TestNewJavaFunctionEventFromPB2(t *testing.T) {
	pbType := pb.InspectorMsgReq_EVENT
	pbFuncType := pb.InspectorMsgReq_Event_FUNC_RETURN
	pbReq := pb.InspectorMsgReq{
		Type:     &pbType,
		EntityId: proto.String("dummy2"),
		Event: &pb.InspectorMsgReq_Event{
			Type: &pbFuncType,
			FuncReturn: &pb.InspectorMsgReq_Event_FuncReturn{
				Name: proto.String("sayHello2"),
			},
		},
		HasJavaSpecificFields: proto.Int32(1),
		JavaSpecificFields: &pb.InspectorMsgReq_JavaSpecificFields{
			ThreadName:           proto.String("threadfoobar"),
			NrParams:             proto.Int32(0),
			Params:               []*pb.InspectorMsgReq_JavaSpecificFields_Params{},
			NrStackTraceElements: proto.Int32(0),
			StackTraceElements:   []*pb.InspectorMsgReq_JavaSpecificFields_StackTraceElement{},
		},
	}

	event, err := NewJavaFunctionEventFromPB(pbReq, time.Now())
	assert.NoError(t, err)
	t.Logf("event: %#v", event)
	javaFunctionEvent, ok := event.(*JavaFunctionEvent)
	if !ok {
		t.Fatal("Cannot convert to JavaFunctionEvent")
	}
	assert.Equal(t, "dummy2", event.EntityID())
	assert.Equal(t, "sayHello2", javaFunctionEvent.FunctionName)
	assert.Equal(t, "threadfoobar", javaFunctionEvent.ThreadName)
	t.Logf("event JSON map: %#v", event.JSONMap())

	action := testDeferredEventDefaultAction(t, event)
	testGOBAction(t, action, event)

	pbAction, pbActionOk := action.(PBAction)
	if !pbActionOk {
		t.Fatalf("non-PBAction action: %#v", action)
	}
	pbRspMsg := pbAction.PBResponseMessage()
	assert.NotNil(t, pbRspMsg, "pbRspMsg should not be nil, action=%#v", action)
	t.Logf("pbRspMsg: %s", pbRspMsg)
}

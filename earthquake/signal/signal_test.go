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
	"bytes"
	"encoding/gob"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/osrg/earthquake/earthquake/util/pb"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	RegisterKnownSignals()
	os.Exit(m.Run())
}

func TestNewLogEventFromJSONString(t *testing.T) {
	s := `
{
    "type": "event",
    "class": "LogEvent",
    "entity": "_earthquake_syslog_inspector",
    "uuid": "1f13eaa6-4b92-45f0-a4de-1236081ec142",
    "deferred": false,
    "option": {
        "src": "zksrv1",
        "message": "foo bar"
    }
}`
	signal, err := NewSignalFromJSONString(s, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	event := signal.(Event)
	t.Logf("event: %#v", event)

	logEvent, ok := event.(*LogEvent)
	if !ok {
		t.Fatal("Cannot convert to LogEvent")
	}

	assert.Equal(t, logEvent.EntityID(), "_earthquake_syslog_inspector")
	assert.Equal(t, logEvent.Deferred(), false)
	assert.Equal(t, logEvent.ID(), "1f13eaa6-4b92-45f0-a4de-1236081ec142")
	assert.Equal(t, logEvent.JSONMap()["option"], map[string]interface{}{
		"src":     "zksrv1",
		"message": "foo bar",
	})
}

func TestNewPacketEventFromJSONString(t *testing.T) {
	s := `
{
    "type": "event",
    "class": "PacketEvent",
    "entity": "_earthquake_ether_inspector",
    "uuid": "1f13eaa6-4b92-45f0-a4de-1236081dc649",
    "deferred": true,
    "option": {
        "src": "zksrv1",
        "dst": "zksrv3",
        "message": {
            "ZkQuorumPacket": {
                "AckQP": {
                    "zxid_low": 0,
                    "zxid_high": 1
                }
            }
        }
    }
}`
	signal, err := NewSignalFromJSONString(s, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	event := signal.(Event)
	t.Logf("event: %#v", event)

	packetEvent, ok := event.(*PacketEvent)
	if !ok {
		t.Fatal("Cannot convert to PacketEvent")
	}

	assert.Equal(t, packetEvent.EntityID(), "_earthquake_ether_inspector")
	assert.Equal(t, packetEvent.Deferred(), true)
	assert.Equal(t, packetEvent.ID(), "1f13eaa6-4b92-45f0-a4de-1236081dc649")
	opt1 := packetEvent.JSONMap()["option"].(map[string]interface{})
	opt2 := map[string]interface{}{
		"src": "zksrv1",
		"dst": "zksrv3",
		"message": map[string]interface{}{
			"ZkQuorumPacket": map[string]interface{}{
				"AckQP": map[string]interface{}{
					"zxid_low":  0,
					"zxid_high": 1,
				},
			},
		},
	}
	assert.Equal(t, opt1["src"], opt2["src"])
	// TODO: compare much more

	action, err := event.DefaultAction()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("default action: %#v", action)
	actionEvent := action.Event()
	assert.Equal(t, event.ID(), actionEvent.ID())

	// try encode
	gobEncodeBuf := bytes.Buffer{}
	gobEncoder := gob.NewEncoder(&gobEncodeBuf)
	if err = gobEncoder.Encode(&action); err != nil {
		t.Fatal(err)
	}
	gobEncoded := gobEncodeBuf.Bytes()

	// try decode
	gobDecoder := gob.NewDecoder(bytes.NewBuffer(gobEncoded))
	var gobDecodedAction Action
	if err = gobDecoder.Decode(&gobDecodedAction); err != nil {
		t.Fatal(err)
	}

	t.Logf("enc/decoded action: %#v", gobDecodedAction)
	assert.Equal(t, action.ID(), gobDecodedAction.ID())

	gobDecodedActionEvent := gobDecodedAction.Event()
	t.Logf("enc/decoded event: %#v", gobDecodedActionEvent)
	assert.Equal(t, event.ID(), gobDecodedActionEvent.ID())
}

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
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("event: %#v", event)

	javaFunctionEvent, ok := event.(*JavaFunctionEvent)
	if !ok {
		t.Fatal("Cannot convert to JavaFunctionEvent")
	}

	assert.Equal(t, event.EntityID(), "dummy")
	assert.Equal(t, event.Deferred(), true)
	assert.Equal(t, javaFunctionEvent.FunctionName, "sayHello")
	t.Logf("event JSON map: %#v", event.JSONMap())

	action, err := event.DefaultAction()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("default action: %#v", action)
	actionEvent := action.Event()
	assert.Equal(t, event.ID(), actionEvent.ID())
	pbAction, pbActionOk := action.(PBAction)
	if !pbActionOk {
		t.Fatalf("non-PBAction action: %#v", action)
	}
	pbRspMsg := pbAction.PBResponseMessage()
	if pbRspMsg == nil {
		t.Fatalf("pbRspMsg should not be nil, action=%#v", action)
	}
	t.Logf("pbRspMsg: %s", pbRspMsg)
}

func TestNewEventAcceptanceActionFromJSONString(t *testing.T) {
	s := `
{
    "type": "action",
    "class": "EventAcceptanceAction",
    "entity": "foobar",
    "uuid": "1f13eaa6-4b92-45f0-a4de-1236081e2442",
    "event_uuid": "1f13eaa6-4b92-45f0-a4de-1236081e2445"
}`
	signal, err := NewSignalFromJSONString(s, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	action := signal.(Action)
	t.Logf("action: %#v", action)

	acceptAction, ok := action.(*EventAcceptanceAction)
	if !ok {
		t.Fatal("Cannot convert to EventAcceptanceAction")
	}

	event := action.Event()
	assert.Equal(t, event.ID(), "1f13eaa6-4b92-45f0-a4de-1236081e2445")
	assert.Nil(t, acceptAction.PBResponseMessage())
}

func TestNewBadEventAcceptanceActionFromJSONString(t *testing.T) {
	// bad EventAcceptanceAction, lacks event_uuid
	s := `
{
    "type": "action",
    "class": "EventAcceptanceAction",
    "entity": "foobar",
    "uuid": "1f13eaa6-4b92-45f0-a4de-1236081e2442"
}`
	signal, err := NewSignalFromJSONString(s, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	action := signal.(Action)
	t.Logf("action: %#v", action)
	assert.Nil(t, action.Event())
}

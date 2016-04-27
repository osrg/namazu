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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewPacketEvent(t *testing.T) {
	event, err := NewPacketEvent("foo", "bar", "baz", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, "foo", event.EntityID())
	action := testDeferredEventDefaultAction(t, event)
	faultAction := testDeferredEventDefaultFaultAction(t, event)
	assert.IsType(t, &EventAcceptanceAction{}, action)
	assert.IsType(t, &PacketFaultAction{}, faultAction)
	testGOBAction(t, action, event)
	testGOBAction(t, faultAction, event)
}

func TestNewPacketEventFromJSONString(t *testing.T) {
	s := `
{
    "type": "event",
    "class": "PacketEvent",
    "entity": "_namazu_ether_inspector",
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
	assert.NoError(t, err)
	event := signal.(Event)
	t.Logf("event: %#v", event)

	packetEvent, ok := event.(*PacketEvent)
	if !ok {
		t.Fatal("Cannot convert to PacketEvent")
	}

	assert.Equal(t, "_namazu_ether_inspector", packetEvent.EntityID())
	assert.Equal(t, "1f13eaa6-4b92-45f0-a4de-1236081dc649", packetEvent.ID())
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
	assert.Equal(t, opt2["src"], opt1["src"])
	// TODO: compare much more

	action := testDeferredEventDefaultAction(t, event)
	faultAction := testDeferredEventDefaultFaultAction(t, event)
	assert.IsType(t, &EventAcceptanceAction{}, action)
	assert.IsType(t, &PacketFaultAction{}, faultAction)
	testGOBAction(t, action, event)
	testGOBAction(t, faultAction, event)
}

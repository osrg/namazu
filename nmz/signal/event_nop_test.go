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

func TestNewNopEvent(t *testing.T) {
	event, err := NewNopEvent("foo", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, "foo", event.EntityID())
	action := testNonDeferredEventDefaultAction(t, event)
	testNonDeferredEventDefaultFaultAction(t, event)
	testGOBAction(t, action, event)
}

func TestNewNopEventFromJSONString(t *testing.T) {
	s := `
{
    "type": "event",
    "class": "NopEvent",
    "entity": "_namazu_syslog_inspector",
    "uuid": "1f13eaa6-4b92-45f0-a4de-1236081ec142",
    "deferred": false,
    "option": {
    }
}`
	signal, err := NewSignalFromJSONString(s, time.Now())
	assert.NoError(t, err)
	event := signal.(Event)
	t.Logf("event: %#v", event)

	nopEvent, ok := event.(*NopEvent)
	if !ok {
		t.Fatal("Cannot convert to NopEvent")
	}

	assert.Equal(t, nopEvent.EntityID(), "_namazu_syslog_inspector")
	assert.Equal(t, nopEvent.ID(), "1f13eaa6-4b92-45f0-a4de-1236081ec142")
	testNonDeferredEventDefaultAction(t, nopEvent)
	testNonDeferredEventDefaultFaultAction(t, nopEvent)
}

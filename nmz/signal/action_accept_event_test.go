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
	assert.NoError(t, err)
	action := signal.(Action)
	t.Logf("action: %#v", action)

	acceptAction, ok := action.(*EventAcceptanceAction)
	if !ok {
		t.Fatal("Cannot convert to EventAcceptanceAction")
	}

	event := action.Event()
	assert.Equal(t, "1f13eaa6-4b92-45f0-a4de-1236081e2445", event.ID())
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
	assert.NoError(t, err, "even for bad action, this err should not happen")
	action := signal.(Action)
	t.Logf("action: %#v", action)
	assert.Nil(t, action.Event(), "there should not be an event because the action lacks event_uuid")
}

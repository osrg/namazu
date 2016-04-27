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
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewFilesystemEvent(t *testing.T) {
	event, err := NewFilesystemEvent("foo", PreMkdir, "/bar", map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, "foo", event.EntityID())
	opt1 := event.JSONMap()["option"].(map[string]interface{})
	opt2 := map[string]interface{}{
		"op":   "pre-mkdir",
		"path": "/bar",
	}
	assert.Equal(t, FilesystemOp(opt2["op"].(string)), opt1["op"])
	// is this inequality really expected?
	assert.NotEqual(t, opt2["op"], opt1["op"])
	assert.Equal(t, opt2["path"], opt1["path"])
	action := testDeferredEventDefaultAction(t, event)
	faultAction := testDeferredEventDefaultFaultAction(t, event)
	assert.IsType(t, &EventAcceptanceAction{}, action)
	assert.IsType(t, &FilesystemFaultAction{}, faultAction)
	testGOBAction(t, action, event)
	testGOBAction(t, faultAction, event)
	assert.False(t, action.Equals(faultAction),
		fmt.Sprintf("\n %s \n vs \n %s", action, faultAction))
}

func TestNewFilesystemEventFromJSONString(t *testing.T) {
	s := `
{
    "type": "event",
    "class": "FilesystemEvent",
    "entity": "_namazu_fs_inspector",
    "uuid": "1f13eaa6-4b92-45f0-a4de-1236081dc649",
    "deferred": true,
    "option": {
        "op": "post-read",
        "path": "/dummy"
    }
}`
	signal, err := NewSignalFromJSONString(s, time.Now())
	assert.NoError(t, err)
	event := signal.(Event)
	t.Logf("event: %#v", event)

	fsEvent, ok := event.(*FilesystemEvent)
	if !ok {
		t.Fatal("Cannot convert to FilesystemEvent")
	}

	assert.Equal(t, "_namazu_fs_inspector", fsEvent.EntityID())
	assert.Equal(t, "1f13eaa6-4b92-45f0-a4de-1236081dc649", fsEvent.ID())
	opt1 := fsEvent.JSONMap()["option"].(map[string]interface{})
	opt2 := map[string]interface{}{
		"op":   "post-read",
		"path": "/dummy",
	}
	assert.Equal(t, opt2["op"], opt1["op"])
	assert.Equal(t, opt2["path"], opt1["path"])

	action := testDeferredEventDefaultAction(t, event)
	faultAction := testDeferredEventDefaultFaultAction(t, event)
	assert.IsType(t, &EventAcceptanceAction{}, action)
	assert.IsType(t, &FilesystemFaultAction{}, faultAction)
	testGOBAction(t, action, event)
	testGOBAction(t, faultAction, event)
	assert.False(t, action.Equals(faultAction),
		fmt.Sprintf("\n %s \n vs \n %s", action, faultAction))
}

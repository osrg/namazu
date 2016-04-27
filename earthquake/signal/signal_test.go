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
	"fmt"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	RegisterKnownSignals()
	os.Exit(m.Run())
}

func testDeferredEventCommon(t *testing.T, event Event,
	f func(Event) (Action, error)) Action {
	t.Logf("event: %s", event)
	assert.True(t, event.Deferred())
	action, err := f(event)
	assert.NoError(t, err)
	t.Logf("action: %s", action)
	assert.Equal(t, event.EntityID(), action.EntityID())

	// compare action vs actionEvent
	actionEvent := action.Event()
	assert.Equal(t, event.EntityID(), actionEvent.EntityID())
	assert.Equal(t, event.ID(), actionEvent.ID())
	assert.True(t, actionEvent.Equals(event),
		fmt.Sprintf("\n %s \n vs \n %s", action, actionEvent))

	// compare action vs action2
	action2, err := f(event)
	assert.NoError(t, err)
	assert.True(t, action.Equals(action2),
		fmt.Sprintf("\n %s \n vs \n %s", action, action2))
	return action
}

func testDeferredEventDefaultAction(t *testing.T, event Event) Action {
	return testDeferredEventCommon(t, event,
		func(e Event) (Action, error) {
			return e.DefaultAction()
		})

}

func testDeferredEventDefaultFaultAction(t *testing.T, event Event) Action {
	return testDeferredEventCommon(t, event,
		func(e Event) (Action, error) {
			return e.DefaultFaultAction()
		})
}

func testNonDeferredEventDefaultAction(t *testing.T, event Event) Action {
	assert.False(t, event.Deferred())
	action, err := event.DefaultAction()
	assert.NoError(t, err)
	assert.NotNil(t, action, "default action cannot be nil, although default action can")
	assert.IsType(t, &NopAction{}, action)
	return action
}

func testNonDeferredEventDefaultFaultAction(t *testing.T, event Event) Action {
	assert.False(t, event.Deferred())
	action, err := event.DefaultFaultAction()
	assert.NoError(t, err)
	assert.Nil(t, action)
	return nil
}

func testGOBAction(t *testing.T, action Action, eventExpected Event) Event {
	// try encode
	gobEncodeBuf := bytes.Buffer{}
	gobEncoder := gob.NewEncoder(&gobEncodeBuf)
	err := gobEncoder.Encode(&action)
	assert.NoError(t, err)
	gobEncoded := gobEncodeBuf.Bytes()

	// try decode
	gobDecoder := gob.NewDecoder(bytes.NewBuffer(gobEncoded))
	var gobDecodedAction Action
	err = gobDecoder.Decode(&gobDecodedAction)
	assert.NoError(t, err)

	t.Logf("enc/decoded action: %#v", gobDecodedAction)
	assert.Equal(t, action.ID(), gobDecodedAction.ID())

	if eventExpected != nil {
		gobDecodedActionEvent := gobDecodedAction.Event()
		t.Logf("enc/decoded event: %#v", gobDecodedActionEvent)
		assert.Equal(t, eventExpected.ID(), gobDecodedActionEvent.ID())
		return gobDecodedActionEvent
	} else {
		return nil
	}
}

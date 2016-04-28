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

package queue

import (
	"flag"
	"os"
	"testing"
	"time"

	. "github.com/osrg/namazu/nmz/signal"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	RegisterKnownSignals()
	os.Exit(m.Run())
}

func makeDummyAction(t *testing.T, entityID string, value int) Action {
	action, err := NewNopAction(entityID, nil)
	assert.NoError(t, err)
	m := map[string]interface{}{"value": value}
	action.(*NopAction).SetOption(m)
	return action
}

func getOptionValueOfAction(t *testing.T, action Action) int {
	assert.NotNil(t, action)
	actionOpt := action.JSONMap()["option"].(map[string]interface{})
	value := actionOpt["value"].(int)
	return value
}

func enqueueActions(t *testing.T, queue *ActionQueue, n int) {
	for i := 0; i < n; i++ {
		action := makeDummyAction(t, queue.EntityID, i)
		queue.Put(action)
	}
}

func dequeueAndVerifyActions(t *testing.T, queue *ActionQueue, n int) {
	for i := 0; i < n; i++ {
		assert.False(t, queue.Peeking())
		action := queue.Peek()
		assert.False(t, queue.Peeking())
		value := getOptionValueOfAction(t, action)
		assert.Equal(t, i, value)
		queue.Delete(action.ID())
	}
	assert.Zero(t, queue.Count())
}

func TestRESTQueue(t *testing.T) {
	entityID := "baz"
	queue, err := RegisterNewQueue(entityID)
	assert.NoError(t, err)
	assert.Equal(t, queue, GetQueue(entityID))
	assert.Equal(t, queue.EntityID, entityID)

	n := 16
	go enqueueActions(t, queue, n)
	dequeueAndVerifyActions(t, queue, n)
	err = UnregisterQueue(entityID)
	assert.NoError(t, err)
}

func TestRESTQueueRace(t *testing.T) {
	entityID := "racy"
	queue, err := RegisterNewQueue(entityID)
	assert.NoError(t, err)
	loserDone := make(chan bool)
	winnerDone := make(chan bool)
	loser := func() {
		t.Log("loser starting")
		<-time.After(0 * time.Second)
		t.Log("loser peeking")
		action := queue.Peek()
		assert.Nil(t, action)
		t.Log("loser done (got nil)")
		loserDone <- true
	}
	winner := func() {
		t.Log("winner starting")
		<-time.After(5 * time.Second)
		t.Log("winner peeking")
		action := queue.Peek()
		assert.NotNil(t, action)
		value := getOptionValueOfAction(t, action)
		t.Logf("winner got %d", value)
		assert.Equal(t, value, 42)
		queue.Delete(action.ID())
		t.Log("winner done")
		winnerDone <- true
	}
	enqueuer := func() {
		<-time.After(10 * time.Second)
		action := makeDummyAction(t, entityID, 42)
		queue.Put(action)
	}
	go loser()
	go winner()
	go enqueuer()
	<-loserDone
	<-winnerDone
	assert.Zero(t, queue.Count())
	err = UnregisterQueue(entityID)
	assert.NoError(t, err)
}

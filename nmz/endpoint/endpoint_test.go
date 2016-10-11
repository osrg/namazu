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

package endpoint

import (
	"flag"
	"fmt"
	"github.com/osrg/namazu/nmz/endpoint/rest"
	"github.com/osrg/namazu/nmz/inspector/transceiver"
	"github.com/osrg/namazu/nmz/signal"
	"github.com/osrg/namazu/nmz/util/config"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/osrg/namazu/nmz/util/mockorchestrator"
	restutil "github.com/osrg/namazu/nmz/util/rest"
	testutil "github.com/osrg/namazu/nmz/util/test"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

var (
	localTransceiver transceiver.Transceiver
	restTransceivers []transceiver.Transceiver
)

const (
	maxRESTEntities = 4
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()

	// Instantiate the orchestrator
	cfg, err := config.NewFromString("{\"restPort\":0}", "json")
	if err != nil {
		panic(err)
	}
	actionCh := make(chan signal.Action)
	eventCh, _ := StartAll(actionCh, cfg)
	mockOrc := mockorchestrator.NewMockOrchestrator(eventCh, actionCh)
	mockOrc.Start()
	defer mockOrc.Shutdown()

	// Instantiate the local transceiver
	localTransceiver, err = transceiver.NewTransceiver("local://", "")
	if err != nil {
		panic(err)
	}
	localTransceiver.Start()

	// Pre-allocate REST transceivers for maxEntities.
	// This pre-allocation is needed because there is currently no way to
	// "unregister" a transceiver entity in REST.
	// FIXME: transceivers should be a local variable.
	restTransceivers = make([]transceiver.Transceiver, maxRESTEntities)
	url := fmt.Sprintf("http://127.0.0.1:%d%s", rest.ActualPort, restutil.APIRoot)
	for i := 0; i < maxRESTEntities; i++ {
		var err error
		entityID := fmt.Sprintf("restentity-%d", i)
		restTransceivers[i], err = transceiver.NewTransceiver(url,
			entityID)
		if err != nil {
			panic(err)
		}
		restTransceivers[i].Start()
	}
	os.Exit(m.Run())
}

func TestEndpointWithPacketEvent_1_1(t *testing.T) {
	testEndpointWithPacketEvent(t, 1, 1)
}

func TestEndpointWithPacketEvent_2_2(t *testing.T) {
	testEndpointWithPacketEvent(t, 2, 2)
}

func TestEndpointWithPacketEvent_10_2(t *testing.T) {
	testEndpointWithPacketEvent(t, 10, 2)
}

func TestEndpointWithPacketEvent_256_4(t *testing.T) {
	testEndpointWithPacketEvent(t, 256, 4)
}

func testEndpointWithPacketEvent(t *testing.T, n, entities int) {
	assert.True(t, entities <= maxRESTEntities)
	var wg sync.WaitGroup
	wg.Add(2 * n)
	go func() {
		for i := 0; i < n; i++ {
			var entityID string
			var trans transceiver.Transceiver
			entityNum := i % entities
			if entityNum%2 == 0 {
				t.Logf("Test %d: Sending (using local trans)", i)
				entityID = fmt.Sprintf("localentity-%d", entityNum)
				trans = localTransceiver
			} else {
				t.Logf("Test %d: Sending (using REST trans)", i)
				entityID = fmt.Sprintf("restentity-%d", entityNum)
				trans = restTransceivers[entityNum]
			}
			event := testutil.NewPacketEvent(t, entityID, i)
			actionCh, err := trans.SendEvent(event)
			t.Logf("Test %d: Sent %s", i, event)
			assert.NoError(t, err)
			wg.Done()
			go func(i int) {
				t.Logf("Test %d: Receiving", i)
				action := <-actionCh
				restoredEvent := action.Event()
				t.Logf("Test %d: Received action %s (event: %s)", i, action, restoredEvent)
				assert.Equal(t, event.EntityID(), restoredEvent.EntityID())
				wg.Done()
			}(i)
		}
	}()
	wg.Wait()
	// TODO: clean up transceivers
}

func TestEndpointShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	testEndpointShouldNotBlockWithPacketEvent(t, 10, 2)
}

func testEndpointShouldNotBlockWithPacketEvent(t *testing.T, n, entities int) {
	assert.True(t, entities <= maxRESTEntities)
	actionChs := make(map[int]chan signal.Action)
	for i := 0; i < n; i++ {
		var entityID string
		var trans transceiver.Transceiver
		entityNum := i % entities
		if entityNum%2 == 0 {
			t.Logf("Test %d: Sending (using local trans)", i)
			entityID = fmt.Sprintf("localentity-%d", entityNum)
			trans = localTransceiver
		} else {
			t.Logf("Test %d: Sending (using REST trans)", i)
			entityID = fmt.Sprintf("restentity-%d", entityNum)
			trans = restTransceivers[entityNum]
		}
		event := testutil.NewPacketEvent(t, entityID, i)
		actionCh, err := trans.SendEvent(event)
		t.Logf("Test %d: Sent %s", i, event)
		assert.NoError(t, err)
		actionChs[i] = actionCh
	}

	for i := 0; i < n; i++ {
		actionCh := actionChs[i]
		t.Logf("Test %d: Receiving", i)
		action := <-actionCh
		restoredEvent := action.Event()
		t.Logf("Test %d: Received action %s (event: %s)", i, action, restoredEvent)
	}
}

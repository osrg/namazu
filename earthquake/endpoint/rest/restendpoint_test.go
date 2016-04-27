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

package rest

import (
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/osrg/namazu/nmz/inspector/transceiver"
	"github.com/osrg/namazu/nmz/signal"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/osrg/namazu/nmz/util/mockorchestrator"
	restutil "github.com/osrg/namazu/nmz/util/rest"
	testutil "github.com/osrg/namazu/nmz/util/test"
	"github.com/stretchr/testify/assert"
)

var (
	srv          *httptest.Server
	transceivers []transceiver.Transceiver
)

const (
	maxEntities = 32
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()

	// Instantiate the orchestrator
	// channels are defined in restendpoint.go globally
	orchestratorActionCh = make(chan signal.Action)
	mockOrc := mockorchestrator.NewMockOrchestrator(orchestratorEventCh, orchestratorActionCh)
	mockOrc.Start()
	defer mockOrc.Shutdown()

	// Instantiate the HTTP server
	srv = httptest.NewServer(newRouter())
	defer srv.Close()
	go actionPropagatorRoutine()

	// Pre-allocate transceivers for maxEntities.
	// This pre-allocation is needed because there is currently no way to
	// "unregister" a transceiver entity.
	// FIXME: transceivers should be a local variable.
	transceivers = make([]transceiver.Transceiver, maxEntities)
	url := srv.URL + restutil.APIRoot
	for i := 0; i < maxEntities; i++ {
		var err error
		entityID := fmt.Sprintf("entity-%d", i)
		transceivers[i], err = transceiver.NewTransceiver(url,
			entityID)
		if err != nil {
			panic(err)
		}
		transceivers[i].Start()
	}

	os.Exit(m.Run())
}

func TestRESTEndpointWithPacketEvent_1_1(t *testing.T) {
	testRESTEndpointWithPacketEvent(t, 1, 1)
}

func TestRESTEndpointWithPacketEvent_2_2(t *testing.T) {
	testRESTEndpointWithPacketEvent(t, 2, 2)
}

func TestRESTEndpointWithPacketEvent_10_10(t *testing.T) {
	testRESTEndpointWithPacketEvent(t, 10, 10)
}

func TestRESTEndpointWithPacketEvent_256_16(t *testing.T) {
	testRESTEndpointWithPacketEvent(t, 256, 16)
}

// testing solely restendpoint.go is difficult.
// so we test resttransceiver together here.
func testRESTEndpointWithPacketEvent(t *testing.T, n, entities int) {
	assert.True(t, entities <= maxEntities)
	var wg sync.WaitGroup
	wg.Add(2 * n)
	go func() {
		for i := 0; i < n; i++ {
			entityID := fmt.Sprintf("entity-%d", i%entities)
			event := testutil.NewPacketEvent(t, entityID, i)
			t.Logf("Test %d: Sending", i)
			actionCh, err := transceivers[i%entities].SendEvent(event)
			t.Logf("Test %d: Sent %s", i, event)
			assert.NoError(t, err)
			wg.Done()
			go func(i int) {
				t.Logf("Test %d: Receiving", i)
				action := <-actionCh
				restoredEvent := action.Event()
				t.Logf("Test %d: Received action %s (event: %s)", i, action, restoredEvent)
				// REST Action JSON only contains the cause event UUID.
				// So restoredEvent is type of &NopEvent{}.
				assert.IsType(t, &signal.NopEvent{}, restoredEvent)
				assert.Equal(t, event.EntityID(), restoredEvent.EntityID())
				wg.Done()
			}(i)
		}
	}()
	wg.Wait()
	// TODO: clean up transceivers
}

func TestRESTEndpointShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	testRESTEndpointShouldNotBlockWithPacketEvent(t, 10, 2)
}

func testRESTEndpointShouldNotBlockWithPacketEvent(t *testing.T, n, entities int) {
	assert.True(t, entities <= maxEntities)
	actionChs := make(map[int]chan signal.Action)
	for i := 0; i < n; i++ {
		entityID := fmt.Sprintf("entity-%d", i%entities)
		event := testutil.NewPacketEvent(t, entityID, i)
		t.Logf("Test %d: Sending", i)
		actionCh, err := transceivers[i%entities].SendEvent(event)
		assert.NoError(t, err)
		t.Logf("Test %d: Sent %s", i, event)
		actionChs[i] = actionCh
	}
	for i := 0; i < n; i++ {
		actionCh := actionChs[i]
		t.Logf("Test %d: Receiving", i)
		action := <-actionCh
		restoredEvent := action.Event()
		t.Logf("Test %d: Received action %s (event: %s)", i, action, restoredEvent)
	}
	// TODO: clean up transceivers
}

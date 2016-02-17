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

package local

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/osrg/earthquake/earthquake/signal"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	testutil "github.com/osrg/earthquake/earthquake/util/test"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	os.Exit(m.Run())
}

func newPacketEvent(t *testing.T, entityID string, value int) signal.Event {
	m := map[string]interface{}{"value": value}
	event, err := signal.NewPacketEvent(entityID, entityID, entityID, m)
	if err != nil {
		t.Fatal(err)
	}
	return event
}

func TestLocalEndpointWithPacketEvent_1_1(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 1, 1)
}

func TestLocalEndpointWithPacketEvent_2_2(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 2, 2)
}

func TestLocalEndpointWithPacketEvent_10_10(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 10, 10)
}

func TestLocalEndpointWithPacketEvent_1000_10(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 1000, 10)
}

func testLocalEndpointWithPacketEvent(t *testing.T, n, entities int) {
	assert.Zero(t, n%entities)
	ep := NewLocalEndpoint()
	defer ep.Shutdown()
	orcActionCh := make(chan signal.Action)
	orcEventCh := ep.Start(orcActionCh)
	mockOrc := testutil.NewMockOrchestrator(orcEventCh, orcActionCh)
	mockOrc.Start()
	defer mockOrc.Shutdown()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			entityID := fmt.Sprintf("entity-%d", i%entities)
			event := newPacketEvent(t, entityID, i)
			t.Logf("Test %d: Sending %s", i, event)
			ep.InspectorEventCh <- event
			t.Logf("Test %d: Sent %s", i, event)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			t.Logf("Test %d: Receiving", i)
			action := <-ep.InspectorActionCh
			event := action.Event()
			t.Logf("Test %d: Received action %s (event: %s)", i, action, event)
		}
	}()
	wg.Wait()
}

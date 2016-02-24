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

package orchestrator

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	localep "github.com/osrg/earthquake/earthquake/endpoint/local"
	"github.com/osrg/earthquake/earthquake/explorepolicy"
	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	testutil "github.com/osrg/earthquake/earthquake/util/test"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	explorepolicy.RegisterKnownExplorePolicies()
	os.Exit(m.Run())
}

func newDumbOrchestrator(t *testing.T, collectTrace bool) *Orchestrator {
	cfg, err := config.NewFromString("{\"explorePolicy\":\"dumb\"}", "json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := explorepolicy.CreatePolicy(cfg.GetString("explorePolicy"))
	if err != nil {
		t.Fatal(err)
	}
	policy.LoadConfig(cfg)
	oc := NewOrchestrator(cfg, policy, collectTrace)
	// FIXME: NewOrchestrator() should return err
	assert.NotNil(t, oc)
	return oc
}

func TestOrchestratorWithNopEvent_1_1_1sec(t *testing.T) {
	testOrchestratorWithNopEvent(t, 1, 1, time.Second)
}

func TestOrchestratorWithNopEvent_2_2_1sec(t *testing.T) {
	testOrchestratorWithNopEvent(t, 2, 2, time.Second)
}

func TestOrchestratorWithNopEvent_10_10_1sec(t *testing.T) {
	testOrchestratorWithNopEvent(t, 10, 10, time.Second)
}

func TestOrchestratorWithNopEvent_1000_10_10sec(t *testing.T) {
	testOrchestratorWithNopEvent(t, 1000, 10, 10*time.Second)
}

func testOrchestratorWithNopEvent(t *testing.T, n, entities int, livenessWait time.Duration) {
	oc := newDumbOrchestrator(t, true)
	oc.Start()
	ep := localep.SingletonLocalEndpoint
	for i := 0; i < n; i++ {
		entityID := fmt.Sprintf("entity-%d", i%entities)
		event := testutil.NewNopEvent(t, entityID, i)
		t.Logf("Test %d: Sending %s", i, event)
		ep.InspectorEventCh <- event
		t.Logf("Test %d: Sent %s", i, event)
		// NOTE: we cannot get NopAction in ih.ActionChan,
		// as NopAction implements OrchestratorSideOnly().
	}
	t.Logf("Sleeping (livenessWait) %s", livenessWait)
	time.Sleep(livenessWait)
	trace := oc.Shutdown()
	assert.Equal(t, n, len(trace.ActionSequence))
}

func TestOrchestratorWithPacketEvent_1_1(t *testing.T) {
	testOrchestratorWithPacketEvent(t, 1, 1, true)
}

func TestOrchestratorWithPacketEvent_2_2(t *testing.T) {
	testOrchestratorWithPacketEvent(t, 2, 2, true)
}

func TestOrchestratorWithPacketEvent_10_10(t *testing.T) {
	testOrchestratorWithPacketEvent(t, 10, 10, true)
}

func TestOrchestratorWithPacketEvent_1000_10(t *testing.T) {
	testOrchestratorWithPacketEvent(t, 1000, 10, true)
}

func TestOrchestratorShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	testOrchestratorWithPacketEvent(t, 10, 2, false)
}

func testOrchestratorWithPacketEvent(t *testing.T, n, entities int, concurrent bool) {
	oc := newDumbOrchestrator(t, true)
	oc.Start()
	ep := localep.SingletonLocalEndpoint

	sender := func() {
		for i := 0; i < n; i++ {
			entityID := fmt.Sprintf("entity-%d", i%entities)
			event := testutil.NewPacketEvent(t, entityID, i)
			t.Logf("Test %d: Sending %s", i, event)
			ep.InspectorEventCh <- event
			t.Logf("Test %d: Sent %s", i, event)
		}
	}
	receiver := func() {
		for i := 0; i < n; i++ {
			t.Logf("Test %d: Receiving", i)
			action := <-ep.InspectorActionCh
			event := action.Event()
			t.Logf("Test %d: Received action %s (event: %s)", i, action, event)
		}
	}

	if concurrent {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			sender()
		}()
		go func() {
			defer wg.Done()
			receiver()
		}()
		wg.Wait()
	} else {
		sender()
		receiver()
	}

	trace := oc.Shutdown()

	assert.Equal(t, n, len(trace.ActionSequence))
	t.Logf("Received %d actions", len(trace.ActionSequence))
	lastValueForEnt0 := 0
	for i := 0; i < n; i++ {
		action := trace.ActionSequence[i]
		entityID := action.EntityID()
		event := action.Event()
		assert.Equal(t, entityID, event.EntityID())
		if entityID == "entity-0" {
			t.Logf("event: %s", event)
			opt := event.JSONMap()["option"].(map[string]interface{})
			// value should be 0, 10, 20, ... (when entities==10)
			value := opt["value"].(int)
			if value > 0 {
				assert.Equal(t, lastValueForEnt0+entities, value)
			}
			lastValueForEnt0 = value
		}
	}
}

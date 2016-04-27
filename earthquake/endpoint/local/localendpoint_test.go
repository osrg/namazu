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

	"github.com/osrg/namazu/nmz/signal"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/osrg/namazu/nmz/util/mockorchestrator"
	testutil "github.com/osrg/namazu/nmz/util/test"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	os.Exit(m.Run())
}

func TestLocalEndpointWithPacketEvent_1_1(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 1, 1, true)
}

func TestLocalEndpointWithPacketEvent_2_2(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 2, 2, true)
}

func TestLocalEndpointWithPacketEvent_10_10(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 10, 10, true)
}

func TestLocalEndpointWithPacketEvent_1000_10(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 1000, 10, true)
}

func TestLocalEndpointShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	testLocalEndpointWithPacketEvent(t, 10, 2, false)
}

func testLocalEndpointWithPacketEvent(t *testing.T, n, entities int, concurrent bool) {
	ep := NewLocalEndpoint()
	defer ep.Shutdown()
	orcActionCh := make(chan signal.Action)
	orcEventCh := ep.Start(orcActionCh)
	mockOrc := mockorchestrator.NewMockOrchestrator(orcEventCh, orcActionCh)
	mockOrc.Start()
	defer mockOrc.Shutdown()

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
}

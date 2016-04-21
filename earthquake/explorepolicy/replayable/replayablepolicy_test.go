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

package replayable

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	testutil "github.com/osrg/earthquake/earthquake/util/test"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	os.Exit(m.Run())
}

func TestReplayableWithPacketEvent_10_2(t *testing.T) {
	xTestPolicyWithPacketEvent(t, 10, 2, true)
}

func TestReplayableWithPacketEvent_10_10(t *testing.T) {
	xTestPolicyWithPacketEvent(t, 10, 10, true)
}

func TestReplayableShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	xTestPolicyWithPacketEvent(t, 10, 2, false)
}

func TestReplayableShouldNotBlockWithPacketEvent_10_10(t *testing.T) {
	xTestPolicyWithPacketEvent(t, 10, 10, false)
}

func newPolicy(t *testing.T, seed string) *Replayable {
	policy := New()
	cfg := config.New()
	cfg.Set("explorePolicy", "replayable")
	cfg.Set("explorePolicyParam", map[string]interface{}{
		"maxInterval": 1 * time.Second,
		"seed":        seed,
	})
	err := policy.LoadConfig(cfg)
	assert.NoError(t, err)
	return policy
}

func xTestPolicyWithPacketEvent(t *testing.T, n, entities int, concurrent bool) {
	seed := "foobar"
	policy := newPolicy(t, seed)
	sender := func() {
		for i := 0; i < n; i++ {
			entityID := fmt.Sprintf("entity-%d", i%entities)
			event := testutil.NewPacketEvent(t, entityID, i).(*signal.PacketEvent)
			hint := fmt.Sprintf("hint-%s-%d", entityID, i)
			event.SetReplayHint(hint)
			t.Logf("Test %d: Sending event (hint=%s)", i, hint)
			policy.QueueEvent(event)
			t.Logf("Test %d: Sent event (hint=%s)", i, hint)
		}
	}
	receiver := func() {
		for i := 0; i < n; i++ {
			t.Logf("Test %d: Receiving", i)
			action := <-policy.ActionChan()
			event := action.Event()
			t.Logf("Test %d: Received action (event hint=%s)",
				i, event.ReplayHint())
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

// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

package explorepolicytester

import (
	"fmt"
	"github.com/osrg/earthquake/earthquake/signal"
	testutil "github.com/osrg/earthquake/earthquake/util/test"
	"sync"
	"testing"
)

// we can't use explorepolicy.ExplorePolicy directly due to import cycle..
type TestableExplorePolicy interface {
	ActionChan() chan signal.Action
	QueueEvent(signal.Event)
}

func XTestPolicyWithPacketEvent(t *testing.T, policy TestableExplorePolicy,
	n, entities int, concurrent bool) {
	sender := func() {
		for i := 0; i < n; i++ {
			entityID := fmt.Sprintf("entity-%d", i%entities)
			event := testutil.NewPacketEvent(t, entityID, i)
			t.Logf("Test %d: Sending %s", i, event)
			policy.QueueEvent(event)
			t.Logf("Test %d: Sent %s", i, event)
		}
	}
	receiver := func() {
		for i := 0; i < n; i++ {
			t.Logf("Test %d: Receiving", i)
			action := <-policy.ActionChan()
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

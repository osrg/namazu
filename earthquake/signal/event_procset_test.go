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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProcSetEvent(t *testing.T) {
	// PIDS need to be string, due to JSON nature
	procs := []string{
		"0", "24", "42",
	}
	event, err := NewProcSetEvent("foo",
		procs,
		map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, "foo", event.EntityID())
	action := testNonDeferredEventDefaultAction(t, event)
	testNonDeferredEventDefaultFaultAction(t, event)
	// not ProcSetSchedAction, as this is an non-deferred event
	assert.IsType(t, &NopAction{}, action)
	testGOBAction(t, action, event)
}

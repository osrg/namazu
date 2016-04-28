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

func TestNewShellAction(t *testing.T) {
	action, err := NewShellAction("true", map[string]interface{}{
		"dummy": "dummy comment string",
	})
	assert.NoError(t, err)
	testGOBAction(t, action, nil)
	orcSide, ok := action.(OrchestratorSideAction)
	assert.True(t, ok)
	assert.True(t, orcSide.OrchestratorSideOnly())
	assert.NoError(t, orcSide.ExecuteOnOrchestrator())
}

func TestNewBadShellAction(t *testing.T) {
	action, err := NewShellAction("false", map[string]interface{}{
		"dummy": "dummy comment string",
	})
	assert.NoError(t, err)
	orcSide, ok := action.(OrchestratorSideAction)
	assert.True(t, ok)
	assert.Error(t, orcSide.ExecuteOnOrchestrator())
}

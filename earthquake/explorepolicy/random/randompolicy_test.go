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

package random

import (
	"flag"
	"fmt"
	"github.com/AkihiroSuda/go-linuxsched"
	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
	tester "github.com/osrg/earthquake/earthquake/util/explorepolicytester"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	os.Exit(m.Run())
}

func newPolicyFromConfigString(s, typ string) (*Random, error) {
	cfg, err := config.NewFromString(s, typ)
	if err != nil {
		panic(err)
	}
	policy := New()
	err = policy.LoadConfig(cfg)
	return policy, err
}

func newPolicy(t *testing.T) *Random {
	cfgTOML := `
explorePolicy = "random"
[explorePolicyParam]
  minInterval = "30ms"
  maxInterval = "100ms"
`
	policy, err := newPolicyFromConfigString(cfgTOML, "toml")
	assert.NoError(t, err)
	return policy
}

func TestRandomPolicyParameters(t *testing.T) {
	defaultConfig := `
explorePolicy = "random"
[explorePolicyParam]
`
	policy, err := newPolicyFromConfigString(defaultConfig, "toml")
	assert.NoError(t, err)
	assert.Equal(t, policy.MinInterval, 0*time.Millisecond)
	assert.Equal(t, policy.MaxInterval, 0*time.Millisecond)
	assert.Empty(t, policy.PrioritizedEntities)
	assert.Zero(t, policy.ShellActionInterval)
	assert.Empty(t, policy.ShellActionCommand)
	assert.True(t, policy.FaultActionProbability < 0.01)
	assert.True(t, policy.ProcResetSchedProbability > 0.09)

	badPolicyNameAllowedForExtensibility := `
explorePolicy = "randomBADBAD"
[explorePolicyParam]
  minInterval = "30ms"
  maxInterval = "100ms"
  prioritizedEntities = ["foo", "bar", "baz"]
  shellActionInterval = "10s"
  shellActionCommand = "echo hello world"
  faultActionProbability = 0.1
  procResetSchedProbability = 0.0
  thisParameterDoesNotExistButShouldNotMatter = 42
`
	policy, err = newPolicyFromConfigString(badPolicyNameAllowedForExtensibility, "toml")
	assert.NoError(t, err)
	assert.Equal(t, policy.MinInterval, 30*time.Millisecond)
	assert.Equal(t, policy.MaxInterval, 100*time.Millisecond)
	// we know this is ugly..
	assert.EqualValues(t, policy.PrioritizedEntities, map[string]bool{"foo": true, "bar": true, "baz": true})
	assert.Equal(t, policy.ShellActionInterval, 10*time.Second)
	assert.Equal(t, policy.ShellActionCommand, "echo hello world")
	assert.True(t, policy.FaultActionProbability > 0.09)
	assert.True(t, policy.ProcResetSchedProbability < 0.01)
}

func TestRandomPolicyWithPacketEvent_10_2(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, newPolicy(t), 10, 2, true)
}

func TestRandomPolicyWithPacketEvent_10_10(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, newPolicy(t), 10, 10, true)
}

func TestRandomPolicyShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, newPolicy(t), 10, 2, false)
}

func TestRandomPolicyShouldNotBlockWithPacketEvent_10_10(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, newPolicy(t), 10, 10, false)
}

func TestRandomPolicyWithProcEvent_100(t *testing.T) {
	testRandomPolicyWithProcEvent(t, 100)
}

func testRandomPolicyWithProcEvent(t *testing.T, n int) {
	cfg := `
explorePolicy = "random"
[explorePolicyParam]
  procResetSchedProbability = 0.5
`
	policy, err := newPolicyFromConfigString(cfg, "toml")
	assert.NoError(t, err)
	normalCount := 0
	deadlineCount := 0
	othersCount := 0
	actionChan := policy.ActionChan()
	procs := make([]string, n)
	for i := 0; i < n; i++ {
		procs[i] = fmt.Sprintf("%d", i)
	}
	event, err := signal.NewProcSetEvent("foo",
		procs,
		map[string]interface{}{})
	assert.NoError(t, err)
	policy.QueueEvent(event)
	action := <-actionChan
	option := action.JSONMap()["option"].(map[string]interface{})
	attrs := option["attrs"].(map[string]linuxsched.SchedAttr)
	for _, attr := range attrs {
		switch attr.Policy {
		case linuxsched.Normal:
			normalCount += 1
		case linuxsched.Deadline:
			deadlineCount += 1
		default:
			othersCount += 1
		}
	}
	t.Logf("normal=%d, deadline=%d, others=%d",
		normalCount, deadlineCount, othersCount)
	assert.True(t, normalCount > 0)
	assert.True(t, deadlineCount > 0)
	assert.True(t, othersCount == 0)
}

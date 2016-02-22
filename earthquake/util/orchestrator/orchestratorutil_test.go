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
	"os"
	"testing"

	"github.com/osrg/earthquake/earthquake/explorepolicy"
	"github.com/osrg/earthquake/earthquake/util/config"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	explorepolicy.RegisterKnownExplorePolicies()
	os.Exit(m.Run())
}

func TestNewAutopilotOrchestrator(t *testing.T) {
	cfg := config.New()
	oc, err := NewAutopilotOrchestrator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, oc)
}

func TestNewBadAutopilotOrchestrator(t *testing.T) {
	tomlString := `
explorePolicy = "this_bad_policy_should_not_exist"
	`
	cfg, err := config.NewFromString(tomlString, "toml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewAutopilotOrchestrator(cfg)
	t.Logf("error is expected here: %s", err)
	assert.Error(t, err)
}

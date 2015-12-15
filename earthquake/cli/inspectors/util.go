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

package inspectors

import (
	"fmt"
	. "github.com/osrg/earthquake/earthquake/explorepolicy"
	. "github.com/osrg/earthquake/earthquake/orchestrator"
	"github.com/osrg/earthquake/earthquake/util/config"
	restutil "github.com/osrg/earthquake/earthquake/util/rest"
)

var defaultOrchestratorURL = fmt.Sprintf("http://localhost:%d%s", restutil.DefaultPort, restutil.APIRoot)

// instantiate new autopilot-mode orchestrator.
//
// autopilot-mode is useful when you are not interested in non-determinism
func NewAutopilotOrchestrator(configFilePath string) (*Orchestrator, error) {
	cfg, err := config.NewFromFile(configFilePath)
	if err != nil {
		return nil, err
	}

	policy, err := CreatePolicy(cfg.GetString("explorePolicy"))
	policy.LoadConfig(cfg)
	orchestrator := NewOrchestrator(cfg, policy, false)
	return orchestrator, nil
}

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

// Package orchestrator provides utilities for orchestrator
package orchestrator

import (
	"github.com/osrg/earthquake/earthquake/explorepolicy"
	"github.com/osrg/earthquake/earthquake/orchestrator"
	"github.com/osrg/earthquake/earthquake/util/config"
)

// URL used for in-binary local orchestrator
const LocalOrchestratorURL = "local://"

// instantiate new autopilot-mode orchestrator.
//
// autopilot-mode is useful when you do not need PB/REST RPC
func NewAutopilotOrchestrator(cfg config.Config) (*orchestrator.Orchestrator, error) {
	policy, err := explorepolicy.CreatePolicy(cfg.GetString("explorePolicy"))
	if err != nil {
		return nil, err
	}
	policy.LoadConfig(cfg)
	oc := orchestrator.NewOrchestrator(cfg, policy, false)
	return oc, nil
}

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
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/util/config"
	. "github.com/osrg/earthquake/earthquake/orchestrator"
	. "github.com/osrg/earthquake/earthquake/explorepolicy"
)

// instantiate new autopilot-mode orchestrator.
//
// autopilot-mode is useful when you are not interested in non-determinism
func NewAutopilotOrchestrator(configFilePath string) (*Orchestrator, error) {
	config, err := ParseConfigFile(configFilePath)
	if err != nil {
		return nil, err
	}

	storageType := config.GetString("storageType")
	if storageType != "" {
		log.Warnf("ignoring storage type %s", storageType)
	}

	policy, err := CreatePolicy(config.GetString("explorePolicy"))
	policy.Init(nil, config.GetStringMap("explorePolicyParam"))
	orchestrator := NewOrchestrator(config, policy)
	return orchestrator, nil
}

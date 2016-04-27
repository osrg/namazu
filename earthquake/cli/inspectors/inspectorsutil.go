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
	"flag"
	"fmt"

	"github.com/osrg/namazu/nmz/util/config"
	ocutil "github.com/osrg/namazu/nmz/util/orchestrator"
)

type commonFlags struct {
	AutopilotConfig string
	OrchestratorURL string
	EntityID        string
}

func initCommon(f *flag.FlagSet, _f *commonFlags, defaultEntityID string) {
	d := fmt.Sprintf("External Earthquake Orchestrator REST endpoint URL (\"%s\" denotes the internal autopilot orchestrator). e.g. http://localhost:10080/api/v3", ocutil.LocalOrchestratorURL)
	f.StringVar(&_f.OrchestratorURL, "orchestrator-url", ocutil.LocalOrchestratorURL, d)

	d = "Entity ID (must be unique in the system)"
	f.StringVar(&_f.EntityID, "entity-id", defaultEntityID, d)

	d = fmt.Sprintf("Path to \"config.toml\" for the internal autopilot orchestrator. Valid only if \"orchestrator_url\" is set to \"%s\".", ocutil.LocalOrchestratorURL)
	f.StringVar(&_f.AutopilotConfig, "autopilot", "", d)
}

func conditionalStartAutopilotOrchestrator(_f commonFlags) (bool, error) {
	var err error
	if _f.OrchestratorURL != ocutil.LocalOrchestratorURL {
		if _f.AutopilotConfig != "" {
			err = fmt.Errorf("external Orchestrator URL is set for the autopilot orchestrator")
		}
		return false, err
	} else {
		var cfg config.Config
		if _f.AutopilotConfig != "" {
			cfg, err = config.NewFromFile(_f.AutopilotConfig)
			if err != nil {
				return false, err
			}
		} else {
			cfg = config.New()
		}
		autopilotOrchestrator, err := ocutil.NewAutopilotOrchestrator(cfg)
		if err != nil {
			return false, err
		}
		go autopilotOrchestrator.Start()
		return true, nil
	}
}

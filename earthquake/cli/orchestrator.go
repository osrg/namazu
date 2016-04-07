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

package cli

import (
	"fmt"

	log "github.com/cihub/seelog"
	mcli "github.com/mitchellh/cli"
	"github.com/osrg/earthquake/earthquake/util/config"
	ocutil "github.com/osrg/earthquake/earthquake/util/orchestrator"
)

// defaultRESTPort is used only if no config is specified.
const defaultRESTPort = 10080

type orchestratorCmd struct {
}

func (cmd orchestratorCmd) Help() string {
	s := `
The orchestrator command just starts the orchestrator.
Basically you should use "earthquake init" and "earthquake orchestrator",
but "earthquake orchestrator" is sometimes useful for interactive operation.

If no config was specified, ` + string(defaultRESTPort) + `is used as a default REST port.
`
	return s
}

func (cmd orchestratorCmd) Run(args []string) int {
	var cfg config.Config
	var err error
	switch len(args) {
	case 0:
		cfg = config.New()
		cfg.Set("restPort", defaultRESTPort)
	case 1:
		cfg, err = config.NewFromFile(args[0])
		if err != nil {
			log.Criticalf("%s", err)
			return 1
		}
	default:
		fmt.Printf("specify <config file path>\n")
		return 1
	}

	orchestrator, err := ocutil.NewAutopilotOrchestrator(cfg)
	if err != nil {
		log.Criticalf("%s", err)
		return 1
	}
	log.Infof("Starting Orchestrator")
	orchestrator.Start()
	for {
	}

	// TODO: support orchestrator.Shutdown() and get the trace
	return 0
}

func (cmd orchestratorCmd) Synopsis() string {
	return "Start the orchestrator without a workspace"
}

func orchestratorCommandFactory() (mcli.Command, error) {
	return orchestratorCmd{}, nil
}

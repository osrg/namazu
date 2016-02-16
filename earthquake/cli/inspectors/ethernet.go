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
	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"
	inspector "github.com/osrg/earthquake/earthquake/inspector/ethernet"
	"github.com/osrg/earthquake/earthquake/util/config"
	ocutil "github.com/osrg/earthquake/earthquake/util/orchestrator"
)

type etherFlags struct {
	AutopilotConfig   string
	OrchestratorURL   string
	EntityID          string
	HookSwitchZMQAddr string
	NFQNumber         int
}

var (
	etherFlagset = flag.NewFlagSet("ethernet", flag.ExitOnError)
	_etherFlags  = etherFlags{}
)

func init() {
	etherFlagset.StringVar(&_etherFlags.OrchestratorURL, "orchestrator-url",
		ocutil.LocalOrchestratorURL, "orchestrator rest url")
	etherFlagset.StringVar(&_etherFlags.EntityID, "entity-id",
		"_earthquake_ethernet_inspector", "Entity ID")
	etherFlagset.StringVar(&_etherFlags.HookSwitchZMQAddr, "hookswitch",
		"ipc:///tmp/earthquake-container-hookswitch-zmq", "HookSwitch ZeroMQ addr")
	etherFlagset.IntVar(&_etherFlags.NFQNumber, "nfq-number",
		-1, "netfilter_queue number")
	etherFlagset.StringVar(&_etherFlags.AutopilotConfig, "autopilot", "",
		"start autopilot-mode orchestrator, if non-empty config path is set")
}

type etherCmd struct {
}

func (cmd etherCmd) Help() string {
	return "Ethernet Inspector"
}

func (cmd etherCmd) Run(args []string) int {
	return runEtherInspector(args)
}

func (cmd etherCmd) Synopsis() string {
	return "Start Ethernet inspector"
}

func EtherCommandFactory() (cli.Command, error) {
	return etherCmd{}, nil
}

func runEtherInspector(args []string) int {
	if err := etherFlagset.Parse(args); err != nil {
		log.Critical(err)
		return 1
	}

	useHookSwitch := _etherFlags.NFQNumber < 0

	if useHookSwitch && _etherFlags.HookSwitchZMQAddr == "" {
		log.Critical("hookswitch is invalid")
		return 1
	}
	if !useHookSwitch && _etherFlags.NFQNumber > 0xFFFF {
		log.Critical("nfq-number is invalid")
		return 1
	}

	if _etherFlags.AutopilotConfig != "" && _etherFlags.OrchestratorURL != ocutil.LocalOrchestratorURL {
		log.Critical("non-default orchestrator url set for autopilot orchestration mode")
		return 1
	}

	if _etherFlags.AutopilotConfig != "" {
		cfg, err := config.NewFromFile(_etherFlags.AutopilotConfig)
		if err != nil {
			panic(log.Critical(err))
		}
		autopilotOrchestrator, err := ocutil.NewAutopilotOrchestrator(cfg)
		if err != nil {
			panic(log.Critical(err))
		}
		log.Info("Starting autopilot-mode orchestrator")
		go autopilotOrchestrator.Start()
	}

	var etherInspector inspector.EthernetInspector
	if useHookSwitch {
		etherInspector = &inspector.HookSwitchInspector{
			OrchestratorURL:   _etherFlags.OrchestratorURL,
			EntityID:          _etherFlags.EntityID,
			HookSwitchZMQAddr: _etherFlags.HookSwitchZMQAddr,
			EnableTCPWatcher:  true,
		}
	} else {
		etherInspector = &inspector.NFQInspector{
			OrchestratorURL:  _etherFlags.OrchestratorURL,
			EntityID:         _etherFlags.EntityID,
			NFQNumber:        uint16(_etherFlags.NFQNumber),
			EnableTCPWatcher: true,
		}
	}

	if err := etherInspector.Start(); err != nil {
		panic(log.Critical(err))
	}

	// NOTREACHED
	return 0
}

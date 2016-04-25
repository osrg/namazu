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

// +build !static

package inspectors

import (
	"flag"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"

	inspector "github.com/osrg/earthquake/earthquake/inspector/ethernet"
)

type etherFlags struct {
	commonFlags
	HookSwitchZMQAddr string
	NFQNumber         int
}

var (
	etherFlagset = flag.NewFlagSet("ethernet", flag.ExitOnError)
	_etherFlags  = etherFlags{}
)

func init() {
	initCommon(etherFlagset, &_etherFlags.commonFlags, "_earthquake_ethernet_inspector")
	etherFlagset.StringVar(&_etherFlags.HookSwitchZMQAddr, "hookswitch",
		"ipc:///tmp/earthquake-container-hookswitch-zmq", "HookSwitch ZeroMQ addr")
	etherFlagset.IntVar(&_etherFlags.NFQNumber, "nfq-number",
		-1, "netfilter_queue number")
}

type etherCmd struct {
}

func EtherCommandFactory() (cli.Command, error) {
	return etherCmd{}, nil
}

func (cmd etherCmd) Help() string {
	return "Please run `earthquake --help inspectors` instead"
}

func (cmd etherCmd) Synopsis() string {
	return "Start Ethernet inspector"
}

func (cmd etherCmd) Run(args []string) int {
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

	autopilot, err := conditionalStartAutopilotOrchestrator(_fsFlags.commonFlags)
	if err != nil {
		log.Critical(err)
		return 1
	}
	log.Infof("Autopilot-mode: %t", autopilot)

	var etherInspector inspector.EthernetInspector
	if useHookSwitch {
		log.Infof("Using hookswitch %s", _etherFlags.HookSwitchZMQAddr)
		etherInspector = &inspector.HookSwitchInspector{
			OrchestratorURL:   _etherFlags.OrchestratorURL,
			EntityID:          _etherFlags.EntityID,
			HookSwitchZMQAddr: _etherFlags.HookSwitchZMQAddr,
			EnableTCPWatcher:  true,
		}
	} else {
		log.Infof("Using NFQ %s", _etherFlags.HookSwitchZMQAddr)
		etherInspector = &inspector.NFQInspector{
			OrchestratorURL:  _etherFlags.OrchestratorURL,
			EntityID:         _etherFlags.EntityID,
			NFQNumber:        uint16(_etherFlags.NFQNumber),
			EnableTCPWatcher: true,
		}
	}

	if err := etherInspector.Serve(); err != nil {
		panic(log.Critical(err))
	}

	// NOTREACHED
	return 0
}

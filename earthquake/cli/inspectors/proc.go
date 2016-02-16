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
	inspector "github.com/osrg/earthquake/earthquake/inspector/proc"
	"github.com/osrg/earthquake/earthquake/util/config"
	ocutil "github.com/osrg/earthquake/earthquake/util/orchestrator"
	"time"
)

type procFlags struct {
	AutopilotConfig string
	OrchestratorURL string
	EntityID        string
	RootPID         int
	WatchInterval   time.Duration
}

var (
	procFlagset = flag.NewFlagSet("proc", flag.ExitOnError)
	_procFlags  = procFlags{}
)

func init() {
	procFlagset.StringVar(&_procFlags.OrchestratorURL, "orchestrator-url", ocutil.LocalOrchestratorURL, "orchestrator rest url")
	procFlagset.StringVar(&_procFlags.EntityID, "entity-id", "_earthquake_proc_inspector", "Entity ID")
	procFlagset.IntVar(&_procFlags.RootPID, "root-pid", -1, "PID for the target process tree")
	procFlagset.DurationVar(&_procFlags.WatchInterval, "watch-interval", 1*time.Second, "Watching interval")
	procFlagset.StringVar(&_procFlags.AutopilotConfig, "autopilot", "",
		"start autopilot-mode orchestrator, if non-empty config path is set")
}

type procCmd struct {
}

func (cmd procCmd) Help() string {
	return "Process Inspector"
}

func (cmd procCmd) Run(args []string) int {
	return runProcInspector(args)
}

func (cmd procCmd) Synopsis() string {
	return "Start process inspector"
}

func ProcCommandFactory() (cli.Command, error) {
	return procCmd{}, nil
}

func runProcInspector(args []string) int {
	if err := procFlagset.Parse(args); err != nil {
		log.Critical(err)
		return 1
	}

	if _procFlags.RootPID <= 0 {
		log.Critical("root-pid is not set (or set to non-positive value)")
		return 1
	}

	if _procFlags.AutopilotConfig != "" && _procFlags.OrchestratorURL != ocutil.LocalOrchestratorURL {
		log.Critical("non-default orchestrator url set for autopilot orchestration mode")
		return 1
	}

	if _procFlags.AutopilotConfig != "" {
		cfg, err := config.NewFromFile(_procFlags.AutopilotConfig)
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
	procInspector := &inspector.ProcInspector{
		OrchestratorURL: _procFlags.OrchestratorURL,
		EntityID:        _procFlags.EntityID,
		RootPID:         _procFlags.RootPID,
		WatchInterval:   _procFlags.WatchInterval,
	}

	if err := procInspector.Start(); err != nil {
		panic(log.Critical(err))
	}

	// NOTREACHED
	return 0
}

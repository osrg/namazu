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
	"time"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"

	inspector "github.com/osrg/earthquake/earthquake/inspector/proc"
)

type procFlags struct {
	commonFlags
	RootPID       int
	WatchInterval time.Duration
}

var (
	procFlagset = flag.NewFlagSet("proc", flag.ExitOnError)
	_procFlags  = procFlags{}
)

func init() {
	initCommon(procFlagset, &_procFlags.commonFlags, "_earthquake_proc_inspector")
	procFlagset.IntVar(&_procFlags.RootPID, "root-pid", -1, "PID for the target process tree")
	procFlagset.DurationVar(&_procFlags.WatchInterval, "watch-interval", 1*time.Second, "Watching interval")
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

	autopilot, err := conditionalStartAutopilotOrchestrator(_fsFlags.commonFlags)
	if err != nil {
		log.Critical(err)
		return 1
	}
	log.Infof("Autopilot-mode: %t", autopilot)

	procInspector := &inspector.ProcInspector{
		OrchestratorURL: _procFlags.OrchestratorURL,
		EntityID:        _procFlags.EntityID,
		RootPID:         _procFlags.RootPID,
		WatchInterval:   _procFlags.WatchInterval,
	}

	if err := procInspector.Serve(); err != nil {
		panic(log.Critical(err))
	}

	// NOTREACHED
	return 0
}

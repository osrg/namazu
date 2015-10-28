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
	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"
	inspector "github.com/osrg/earthquake/earthquake/inspector/fs"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	restutil "github.com/osrg/earthquake/earthquake/util/rest"
	"github.com/osrg/hookfs/hookfs"
)

type fsFlags struct {
	AutopilotConfig string
	OrchestratorURL string
	EntityID        string
	OriginalDir     string
	Mountpoint      string
}

var (
	fsFlagset              = flag.NewFlagSet("fs", flag.ExitOnError)
	_fsFlags               = fsFlags{}
	defaultOrchestratorURL = fmt.Sprintf("http://localhost:%d%s", restutil.DefaultPort, restutil.APIRoot)
)

func init() {
	fsFlagset.StringVar(&_fsFlags.OrchestratorURL, "orchestrator-url", defaultOrchestratorURL, "orchestrator rest url")
	fsFlagset.StringVar(&_fsFlags.EntityID, "entity-id", "_earthquake_fs_inspector", "Entity ID")
	fsFlagset.StringVar(&_fsFlags.OriginalDir, "original-dir", "", "FUSE Original Directory")
	fsFlagset.StringVar(&_fsFlags.Mountpoint, "mount-point", "", "FUSE Mount Point")
	fsFlagset.StringVar(&_fsFlags.AutopilotConfig, "autopilot", "",
		"start autopilot-mode orchestrator, if non-empty config path is set")
}

type fsCmd struct {
}

func (cmd fsCmd) Help() string {
	return "Filesystem Inspector (uses FUSE)"
}

func (cmd fsCmd) Run(args []string) int {
	return runFsInspector(args)
}

func (cmd fsCmd) Synopsis() string {
	return "Start filesystem inspector"
}

func FsCommandFactory() (cli.Command, error) {
	return fsCmd{}, nil
}

func runFsInspector(args []string) int {
	if err := fsFlagset.Parse(args); err != nil {
		log.Critical(err)
		return 1
	}

	if _fsFlags.OriginalDir == "" {
		log.Critical("original-dir is not set")
		return 1
	}

	if _fsFlags.Mountpoint == "" {
		log.Critical("mount-point is not set")
		return 1
	}

	if _fsFlags.AutopilotConfig != "" && _fsFlags.OrchestratorURL != defaultOrchestratorURL {
		log.Critical("non-default orchestrator url set for autopilot orchestration mode")
		return 1
	}

	if logutil.Debug {
		// log level: 0..2
		hookfs.SetLogLevel(1)
	} else {
		hookfs.SetLogLevel(0)
	}

	if _fsFlags.AutopilotConfig != "" {
		autopilotOrchestrator, err := NewAutopilotOrchestrator(_fsFlags.AutopilotConfig)
		if err != nil {
			panic(log.Critical(err))
		}
		log.Info("Starting autopilot-mode orchestrator")
		go autopilotOrchestrator.Start()
	}

	hook := &inspector.EQFSHook{
		OrchestratorURL: _fsFlags.OrchestratorURL,
		EntityID:        _fsFlags.EntityID,
	}

	fs, err := hookfs.NewHookFs(_fsFlags.OriginalDir, _fsFlags.Mountpoint, hook)
	if err != nil {
		panic(log.Critical(err))
	}
	log.Infof("Serving %s", fs)
	log.Infof("Please run `fusermount -u %s` after using this, manually", _fsFlags.Mountpoint)
	if err = fs.Serve(); err != nil {
		panic(log.Critical(err))
	}
	// NOTREACHED
	return 0
}

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
	"github.com/osrg/hookfs/hookfs"

	inspector "github.com/osrg/namazu/nmz/inspector/fs"
	logutil "github.com/osrg/namazu/nmz/util/log"
)

type fsFlags struct {
	commonFlags
	OriginalDir string
	Mountpoint  string
}

var (
	fsFlagset = flag.NewFlagSet("fs", flag.ExitOnError)
	_fsFlags  = fsFlags{}
)

func init() {
	initCommon(fsFlagset, &_fsFlags.commonFlags, "_namazu_fs_inspector")
	fsFlagset.StringVar(&_fsFlags.OriginalDir, "original-dir", "", "FUSE Original Directory")
	fsFlagset.StringVar(&_fsFlags.Mountpoint, "mount-point", "", "FUSE Mount Point")
}

type fsCmd struct {
}

func FsCommandFactory() (cli.Command, error) {
	return fsCmd{}, nil
}

func (cmd fsCmd) Help() string {
	return "Please run `namazu --help inspectors` instead"
}

func (cmd fsCmd) Synopsis() string {
	return "Start filesystem inspector"
}

func (cmd fsCmd) Run(args []string) int {
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

	autopilot, err := conditionalStartAutopilotOrchestrator(_fsFlags.commonFlags)
	if err != nil {
		log.Critical(err)
		return 1
	}
	log.Infof("Autopilot-mode: %t", autopilot)

	if logutil.Debug {
		// log level: 0..2
		hookfs.SetLogLevel(1)
	} else {
		hookfs.SetLogLevel(0)
	}

	hook := &inspector.FilesystemInspector{
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

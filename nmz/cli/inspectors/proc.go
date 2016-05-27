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
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"

	inspector "github.com/osrg/namazu/nmz/inspector/proc"
)

type procFlags struct {
	commonFlags
	RootPID       int
	WatchInterval time.Duration
	Cmd           string
	Stdout        string
	Stderr        string
}

var (
	procFlagset = flag.NewFlagSet("proc", flag.ExitOnError)
	_procFlags  = procFlags{}
)

func init() {
	initCommon(procFlagset, &_procFlags.commonFlags, "_namazu_proc_inspector")
	procFlagset.IntVar(&_procFlags.RootPID, "pid", -1, "PID for the target process tree")
	procFlagset.DurationVar(&_procFlags.WatchInterval, "watch-interval", 1*time.Second, "Watching interval")
	procFlagset.StringVar(&_procFlags.Cmd, "cmd", "", "Command for target process")
	procFlagset.StringVar(&_procFlags.Stdout, "stdout", "", "Stdout for target process (used if -cmd option is given)")
	procFlagset.StringVar(&_procFlags.Stderr, "stderr", "", "Stderr for target process (used if -cmd option is given)")
}

type procCmd struct {
}

func ProcCommandFactory() (cli.Command, error) {
	return procCmd{}, nil
}

func (cmd procCmd) Help() string {
	return "Please run `nmz --help inspectors` instead"
}

func (cmd procCmd) Synopsis() string {
	return "Start process inspector"
}

func (cmd procCmd) Run(args []string) int {
	if err := procFlagset.Parse(args); err != nil {
		log.Critical(err)
		return 1
	}

	pid := _procFlags.RootPID
	endCh := make(chan struct{})

	if pid <= 0 {
		if _procFlags.Cmd != "" {
			args := strings.Split(_procFlags.Cmd, " ")
			cmd := exec.Command(args[0], args[1:]...)

			if _procFlags.Stdout == "" {
				cmd.Stdout = os.Stdout
			} else {
				f, err := os.OpenFile(_procFlags.Stdout, os.O_WRONLY|os.O_CREATE, 0622)
				if err != nil {
					log.Critical("failed to open a file %s for stdout: %s", _procFlags.Stdout, err)
					return 1
				}
				cmd.Stdout = f
			}

			if _procFlags.Stderr == "" {
				cmd.Stderr = os.Stderr
			} else {
				f, err := os.OpenFile(_procFlags.Stderr, os.O_WRONLY|os.O_CREATE, 0622)
				if err != nil {
					log.Critical("failed to open a file %s for stderr: %s", _procFlags.Stderr, err)
					return 1
				}
				cmd.Stderr = f
			}

			err := cmd.Start()
			if err != nil {
				log.Critical("failed to cmd.Start: %s", err)
				return 1
			}

			pid = cmd.Process.Pid

			go func() {
				err := cmd.Wait()
				if err != nil {
					log.Critical("failed to cmd.Wait: %s", err)
				}
				endCh <- struct{}{}
			}()
		} else {
			log.Critical("pid and command line are not set (or set to non-positive value)")
			return 1
		}
	} else if _procFlags.Cmd != "" {
		log.Critical("you cannot set both pid and command line")
		return 1
	}

	autopilot, err := conditionalStartAutopilotOrchestrator(_procFlags.commonFlags)
	if err != nil {
		log.Critical(err)
		return 1
	}
	log.Infof("Autopilot-mode: %t", autopilot)

	procInspector := &inspector.ProcInspector{
		OrchestratorURL: _procFlags.OrchestratorURL,
		EntityID:        _procFlags.EntityID,
		RootPID:         pid,
		WatchInterval:   _procFlags.WatchInterval,
	}

	if err := procInspector.Serve(endCh); err != nil {
		panic(log.Critical(err))
	}

	// NOTREACHED
	return 0
}

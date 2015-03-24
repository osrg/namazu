// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

package main

import (
	"flag"
	"fmt"
	"os"
)

type orchestratorFlags struct {
	Debug             bool
	ExecutionFilePath string
	LogFilePath       string
	ListenTCPPort     int
	Daemonize         bool

	SearchMode        bool
	InitSearchModeDir bool
	SearchModeDir     string
}

type guestAgentFlags struct {
	VirtIOPathIn  string
	VirtIOPathOut string
	Debug         bool
	Daemonize     bool
	LogFilePath   string
	MachineID     string
	ListenTCPPort int
}

var (
	globalFlagset = flag.NewFlagSet("earthquake", flag.ExitOnError)
	globalFlags   = struct {
		LaunchOrchestrator bool
		LaunchGuestAgent   bool
		SearchTools        string
	}{}

	orchestratorFlagset = flag.NewFlagSet("orchestrator", flag.ExitOnError)
	_orchestratorFlags  = orchestratorFlags{}

	guestAgentFlagset = flag.NewFlagSet("guestagent", flag.ExitOnError)
	_guestAgentFlags  = guestAgentFlags{}
)

func init() {
	globalFlagset.BoolVar(&globalFlags.LaunchOrchestrator, "launch-orchestrator", false, "Launch orchestrator")
	globalFlagset.BoolVar(&globalFlags.LaunchGuestAgent, "launch-guestagent", false, "Launch guestagent")
	globalFlagset.StringVar(&globalFlags.SearchTools, "search-tools", "", "Tools related to search mode")

	orchestratorFlagset.BoolVar(&_orchestratorFlags.Debug, "debug", false, "Debug mode")
	orchestratorFlagset.StringVar(&_orchestratorFlags.ExecutionFilePath, "execution-file-path", "", "Path of execution file")
	orchestratorFlagset.StringVar(&_orchestratorFlags.LogFilePath, "log-file-path", "", "Path of log file")
	orchestratorFlagset.IntVar(&_orchestratorFlags.ListenTCPPort, "listen-tcp-port", 10000, "TCP Port to listen on")
	orchestratorFlagset.BoolVar(&_orchestratorFlags.Daemonize, "daemonize", false, "Daemonize")
	orchestratorFlagset.BoolVar(&_orchestratorFlags.SearchMode, "search-mode", false, "Run search mode")
	orchestratorFlagset.BoolVar(&_orchestratorFlags.InitSearchModeDir, "init-search-mode-directory", false, "Initialize a directory for storing search mode info")
	orchestratorFlagset.StringVar(&_orchestratorFlags.SearchModeDir, "search-mode-directory", "", "Directory which has config and pas traces of search mode testing")

	guestAgentFlagset.StringVar(&_guestAgentFlags.VirtIOPathIn, "virtio-path-in", "", "Path of VirtIO channel (input side)")
	guestAgentFlagset.StringVar(&_guestAgentFlags.VirtIOPathOut, "virtio-path-out", "", "Path of VirtIO channel (output side)")
	guestAgentFlagset.BoolVar(&_guestAgentFlags.Debug, "debug", false, "Debug mode")
	guestAgentFlagset.BoolVar(&_guestAgentFlags.Daemonize, "daemonize", true, "Daemonize")
	guestAgentFlagset.StringVar(&_guestAgentFlags.LogFilePath, "log-file-path", "/var/log/earthquake.log", "Path of log file")
	guestAgentFlagset.StringVar(&_guestAgentFlags.MachineID, "machine-id", "", "ID of this machine")
	guestAgentFlagset.IntVar(&_guestAgentFlags.ListenTCPPort, "listen-tcp-port", 10000, "TCP Port to listen on")
}

func main() {
	if len(os.Args) < 2 {
		// todo: print usage
		fmt.Printf("specify orchestrator or guestagent for launching\n")
		os.Exit(1)
	}
	globalArgs := []string{os.Args[1]}

	globalFlagset.Parse(globalArgs)

	if globalFlags.SearchTools != "" {
		runSearchTools(globalFlags.SearchTools, os.Args[2:])
		return
	}

	if globalFlags.LaunchOrchestrator && globalFlags.LaunchGuestAgent {
		fmt.Printf("don't specify both of orchestrator and guestagent for launching\n")
		os.Exit(1)
	}

	if globalFlags.LaunchOrchestrator {
		orchestratorFlagset.Parse(os.Args[2:])
		launchOrchestrator(_orchestratorFlags)
		return
	}

	if globalFlags.LaunchGuestAgent {
		guestAgentFlagset.Parse(os.Args[2:])
		launchGuestAgent(_guestAgentFlags)
		return
	}

	fmt.Printf("specify orchestrator or guestagent for launching\n")
	os.Exit(1)
}

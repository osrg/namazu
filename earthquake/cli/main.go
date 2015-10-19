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
	"os"

	log "github.com/cihub/seelog"
	mcli "github.com/mitchellh/cli"
	. "github.com/osrg/earthquake/earthquake/explorepolicy"
	. "github.com/osrg/earthquake/earthquake/signal"
	. "github.com/osrg/earthquake/earthquake/util/log"
)

func cliInit(debug bool) {
	InitLog("", debug)
	RegisterKnownSignals()
	RegisterKnownExplorePolicies()
}

const EarthquakeVersion = "0.1.2"

func recoverer(debug bool) {
	if r := recover(); r != nil {
		log.Criticalf("PANIC: %s", r)
		if debug {
			panic(r)
		} else {
			log.Info("Hint: For debug info, please set \"EQ_DEBUG\" to 1.")
			os.Exit(1)
		}
	}
}

func CLIMain(args []string) int {
	debug := os.Getenv("EQ_DEBUG") != ""
	cliInit(debug)
	defer recoverer(debug)
	c := mcli.NewCLI(args[0], EarthquakeVersion)
	c.Args = args[1:]
	c.Commands = map[string]mcli.CommandFactory{
		"init":       initCommandFactory,
		"run":        runCommandFactory,
		"tools":      toolsCommandFactory,
		"inspectors": inspectorsCommandFactory,
	}
	exitStatus, err := c.Run()
	if err != nil {
		fmt.Printf("failed to execute command: %s\n", err)
	}
	return exitStatus
}

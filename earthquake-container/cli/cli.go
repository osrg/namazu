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
	eqcli "github.com/osrg/earthquake/earthquake/cli"
	"os"
)

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
	eqcli.CLIInit(debug)
	defer recoverer(debug)
	if len(args) < 2 {
		fmt.Printf("Usage: %s [OPTIONS] COMMAND [arg...]\n", args[0])
		fmt.Printf("\n")
		fmt.Printf("Docker Container + Earthquake Testing Framework\n")
		fmt.Printf("\n")
		fmt.Printf("Commands:\n")
		fmt.Printf("\trun\tRun a command in a new container\n")
		fmt.Printf("\n")
		return 0
	}
	switch args[1] {
	case "run":
		return run(args[1:])
	}
	fmt.Fprintf(os.Stderr, "'%s' is not a earthquake-container command.", args[1])
	return 1
}

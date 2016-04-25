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

	mcli "github.com/mitchellh/cli"
	crun "github.com/osrg/earthquake/earthquake/cli/container/run"
)

type containerCmd struct {
}

func (cmd containerCmd) Help() string {
	return "Not documented yet."
}

func (cmd containerCmd) Run(args []string) int {
	if len(args) < 1 {
		fmt.Printf("Usage: earthquake container run [OPTIONS] COMMAND [arg...]\n")
		fmt.Printf("\n")
		fmt.Printf("Docker Container + Earthquake Testing Framework\n")
		fmt.Printf("\n")
		fmt.Printf("Commands:\n")
		fmt.Printf("\trun\tRun a command in a new container\n")
		fmt.Printf("\n")
		return 0
	}
	switch args[0] {
	case "run":
		return crun.Run(args)
	}
	fmt.Fprintf(os.Stderr, "'%s' is not a earthquake-container command.\n", args[0])
	return 1
}

func (cmd containerCmd) Synopsis() string {
	return "Docker-like CLI"
}

func containerCommandFactory() (mcli.Command, error) {
	return containerCmd{}, nil
}

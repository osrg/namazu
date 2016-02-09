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

	mcli "github.com/mitchellh/cli"
	"github.com/osrg/earthquake/earthquake/cli/inspectors"
)

type inspectorsCmd struct {
}

func (cmd inspectorsCmd) Help() string {
	return `
	Earthquake Inspectors
	- proc:     Process inspector
	- fs:       Filesystem inspector
	- ethernet: Ethernet inspector

	NOTE: this binary does NOT include following inspectors:
	- Java Inspector:     (included in earthquake/inspector/java)
	- C Inspector:        (included in earthquake/inspector/c)
	`
}

func (cmd inspectorsCmd) Run(args []string) int {
	c := mcli.NewCLI("earthquake inspectors", EarthquakeVersion)
	c.Args = args
	c.Commands = map[string]mcli.CommandFactory{
		"proc":     inspectors.ProcCommandFactory,
		"fs":       inspectors.FsCommandFactory,
		"ethernet": inspectors.EtherCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Printf("failed to execute inspector: %s\n", err)
	}

	return exitStatus
}

func (cmd inspectorsCmd) Synopsis() string {
	return "Start inspectors"
}

func inspectorsCommandFactory() (mcli.Command, error) {
	return inspectorsCmd{}, nil
}

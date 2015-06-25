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

package main

import (
	"fmt"
	"github.com/mitchellh/cli"

	"./searchtools"
)

type searchToolsCmd struct {
}

func (cmd searchToolsCmd) Help() string {
	return "searchtools help (todo)"
}

func (cmd searchToolsCmd) Run(args []string) int {
	c := cli.NewCLI("earthquake search tools", "0.0.0")
	c.Args = args
	c.Commands = map[string]cli.CommandFactory{
		"calc-dup":   searchtools.CalcDupCommandFactory,
		"visualize":  searchtools.VisualizeCommandFactory,
		"dump-trace": searchtools.DumpTraceCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Printf("failed to execute search tool: %s\n", err)
	}

	return exitStatus
}

func (cmd searchToolsCmd) Synopsis() string {
	return "searchtools subcommand"
}

func searchToolsCommandFactory() (cli.Command, error) {
	return searchToolsCmd{}, nil
}

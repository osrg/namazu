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
	coreutil "github.com/osrg/namazu/nmz/util/core"
)

func help() string {
	s := `
Basically the inspectors command is enough for getting started.
If you want to run a specific test scenario repeatedly, the init command and the run command are useful.

For further information, please visit the web site: https://github.com/osrg/namazu
`
	return s
}

func CLIMain(args []string) int {
	coreutil.Init()
	defer coreutil.Recoverer()
	c := mcli.NewCLI(args[0], coreutil.NamazuVersion)
	c.Args = args[1:]
	c.Commands = map[string]mcli.CommandFactory{
		"init":         initCommandFactory,
		"run":          runCommandFactory,
		"orchestrator": orchestratorCommandFactory,
		"inspectors":   inspectorsCommandFactory,
		"tools":        toolsCommandFactory,
		"container":    containerCommandFactory,
	}
	c.HelpFunc = func(commands map[string]mcli.CommandFactory) string {
		s := (mcli.BasicHelpFunc(args[0]))(commands)
		s += help()
		return s
	}
	exitStatus, err := c.Run()
	if err != nil {
		fmt.Printf("failed to execute command: %s\n", err)
	}
	return exitStatus
}

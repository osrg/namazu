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
	"github.com/osrg/namazu/nmz/cli/inspectors"
	coreutil "github.com/osrg/namazu/nmz/util/core"
)

type inspectorsCmd struct {
}

func (cmd inspectorsCmd) Help() string {
	// FIXME: much more helpful help string
	return `
The inspectors command starts an Namazu inspector.

If -orchestrator-url is set, the inspector connects the external orchestrator.
For how to start the external orchestrator, please refer to the help of the run command.
(earthquake --help run)

Note that you have to set -entity-id to an unique value if you connect multiple inspectors to the external orchestrator.

If -orchestrator-url is not set, the inspector connects the embedded orchestrator.
You can specify the configuration file for the embedded orchestrator by setting -autopilot <config.toml>.


Process inspector (proc)
    Inspects running Linux process information, and set scheduling attributes.

    Typical usage: earthquake inspectors proc -root-pid 42 -watch-interval 1s

    Event signals: ProcSetEvent
    Action signals: ProcSetSchedAction


Filesystem inspector (fs)
    Inspects file access information, and inject delays and faults.
    Implemented in FUSE.

    Typical usage: earthquake inspectors fs -original-dir /tmp/eqfs-orig -mount-point /tmp/eqfs

    Event signals: FilesystemEvent
    Action signals: EventAcceptanceAction, FilesystemFaultAction


Ethernet inspector (ethernet)
    Inspects Ethernet packet information, and inject delays and faults.
    Implemented in Linux netfilter / Openflow.
    For Openflow implementation, you have to install hookswitch: https://github.com/osrg/hookswitch

    Typical usage: earthquake inspectors ethernet -nfq-number 42

    Event signals: PacketEvent
    Action signals: EventAcceptanceAction, PacketFaultAction


NOTE: this binary does NOT include the following inspectors:
    Java Inspector:     (included in misc/inspector/java)
    C Inspector:        (included in misc/inspector/c, NOT MAINTAINED)

NOTE: Python implementation for Ethernet inspector is also available in misc/pynmz.
You can also implement your own inspector in an arbitrary language.
`
}

func (cmd inspectorsCmd) Run(args []string) int {
	c := mcli.NewCLI("earthquake inspectors", coreutil.NamazuVersion)
	c.Args = args
	c.Commands = map[string]mcli.CommandFactory{
		"proc":     inspectors.ProcCommandFactory,
		"fs":       inspectors.FsCommandFactory,
		"ethernet": inspectors.EtherCommandFactory,
	}
	c.HelpFunc = func(commands map[string]mcli.CommandFactory) string {
		s := (mcli.BasicHelpFunc("earthquake inspectors"))(commands)
		s += cmd.Help()
		return s
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

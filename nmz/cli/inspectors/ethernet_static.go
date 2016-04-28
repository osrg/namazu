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

// +build static

package inspectors

import (
	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"

	coreutil "github.com/osrg/namazu/nmz/util/core"
)

type etherCmd struct {
}

func EtherCommandFactory() (cli.Command, error) {
	return etherCmd{}, nil
}

func (cmd etherCmd) Help() string {
	return "Please run `nmz --help inspectors` instead"
}

func (cmd etherCmd) Synopsis() string {
	return "Start Ethernet inspector"
}

func (cmd etherCmd) Run(args []string) int {
	log.Critical(coreutil.EthernetInspectorNotBuiltErr)
	return 1
}

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

package container

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/cihub/seelog"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/osrg/namazu/nmz/container/ns"
)

func SetupNFQUEUE(c *docker.Container, queueNum int, hookInput bool, disableBypass bool) error {
	err := ns.EnterDockerNetNs(c)
	if err != nil {
		return err
	}
	defer ns.LeaveNetNs()

	chain := "OUTPUT"
	if hookInput {
		chain = "INPUT"
	}
	iptArg := []string{"-A", chain, "-j", "NFQUEUE", "--queue-num", fmt.Sprintf("%d", queueNum)}
	if !disableBypass {
		iptArg = append(iptArg, "--queue-bypass")
	}

	log.Debugf("Running `iptables` with %s", iptArg)
	cmd := exec.Command("iptables", iptArg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

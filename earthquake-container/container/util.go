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
	log "github.com/cihub/seelog"
	docker "github.com/fsouza/go-dockerclient"
	cliinsputil "github.com/osrg/earthquake/earthquake/cli/inspectors"
	"github.com/osrg/earthquake/earthquake/inspector/ethernet"
	"github.com/vishvananda/netns"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func NewDockerClient() (*docker.Client, error) {
	host := os.Getenv("DOCKER_HOST")
	hostIsLocal := host == "" || strings.HasPrefix(host, "unix://")
	if !hostIsLocal {
		log.Warnf("Detected DOCKER_HOST %s. This should not be remote.",
			host)
	}
	return docker.NewClientFromEnv()
}

var origNs netns.NsHandle
var newNs netns.NsHandle

func enterDockerNetNs(container *docker.Container) error {
	var err error
	origNs, err = netns.Get()
	if err != nil {
		return log.Criticalf("could not obtain current netns. netns not supported?: %s", err)
	}
	// netns.GetFromDocker() does not work for recent dockers.
	// So we use netns.GetFromPid() directly.
	// https://github.com/vishvananda/netns/pull/10
	newNs, err = netns.GetFromPid(container.State.Pid)
	if err != nil {
		return log.Criticalf("Could not get netns for container %s (pid=%d). The container has exited??: %s", container.ID, container.State.Pid, err)
	}
	runtime.LockOSThread()
	netns.Set(newNs)
	return nil
}

func leaveNetNs() {
	netns.Set(origNs)
	runtime.UnlockOSThread()
	origNs.Close()
	newNs.Close()
}

func SetupNFQUEUE(container *docker.Container, queueNum int, hookInput bool, disableBypass bool) error {
	err := enterDockerNetNs(container)
	if err != nil {
		return err
	}
	defer leaveNetNs()

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

func StartOrchestrator(path string) error {
	// TODO: refactor. should not use cliinsputil.
	autopilotOrchestrator, err := cliinsputil.NewAutopilotOrchestrator(path)
	if err != nil {
		return err
	}
	autopilotOrchestrator.Start()
	return nil
}

func StartEthernetInspector(container *docker.Container, queueNum int) error {
	err := enterDockerNetNs(container)
	if err != nil {
		return err
	}
	// FIXME: should not use REST RPC for in-process communication
	insp := &ethernet.NFQInspector{
		OrchestratorURL:  "http://localhost:10080/api/v3",
		EntityID:         "_earthquake_ethernet_inspector",
		NFQNumber:        uint16(queueNum),
		EnableTCPWatcher: true,
	}
	defer leaveNetNs()
	insp.Start()
	return nil
}

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

package core

import (
	docker "github.com/fsouza/go-dockerclient"
	"github.com/osrg/earthquake/earthquake-container/container"
	"github.com/osrg/earthquake/earthquake/inspector/ethernet"
	"github.com/osrg/earthquake/earthquake/inspector/proc"
	"github.com/osrg/earthquake/earthquake/util/config"
	ocutil "github.com/osrg/earthquake/earthquake/util/orchestrator"
	"time"
)

func StartOrchestrator(cfg config.Config) error {
	autopilotOrchestrator, err := ocutil.NewAutopilotOrchestrator(cfg)
	if err != nil {
		return err
	}
	autopilotOrchestrator.Start()
	return nil
}

func StartEthernetInspector(c *docker.Container, queueNum int) error {
	err := container.EnterDockerNetNs(c)
	if err != nil {
		return err
	}
	insp := &ethernet.NFQInspector{
		OrchestratorURL:  ocutil.LocalOrchestratorURL,
		EntityID:         "_earthquake_container_ethernet_inspector",
		NFQNumber:        uint16(queueNum),
		EnableTCPWatcher: true,
	}
	defer container.LeaveNetNs()
	insp.Start()
	return nil
}

func StartProcInspector(c *docker.Container, watchInterval time.Duration) error {
	insp := &proc.ProcInspector{
		OrchestratorURL: ocutil.LocalOrchestratorURL,
		EntityID:        "_earthquake_container_proc_inspector",
		RootPID:         c.State.Pid,
		WatchInterval:   watchInterval,
	}
	insp.Start()
	return nil
}

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

// +build !static

package container

import (
	docker "github.com/fsouza/go-dockerclient"

	"github.com/osrg/namazu/nmz/container/ns"
	"github.com/osrg/namazu/nmz/inspector/ethernet"
	ocutil "github.com/osrg/namazu/nmz/util/orchestrator"
)

func ServeEthernetInspector(c *docker.Container, queueNum int) error {
	err := ns.EnterDockerNetNs(c)
	if err != nil {
		return err
	}
	insp := &ethernet.NFQInspector{
		OrchestratorURL:  ocutil.LocalOrchestratorURL,
		EntityID:         "_namazu_container_ethernet_inspector",
		NFQNumber:        uint16(queueNum),
		EnableTCPWatcher: true,
	}
	defer ns.LeaveNetNs()
	insp.Serve()
	// NOTREACHED
	return nil
}

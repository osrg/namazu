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
	"github.com/osrg/namazu/nmz/util/config"
)

func StartEarthquakeRoutines(c *docker.Container, cfg config.Config) error {
	log.Debugf("Starting Orchestrator")
	go func() {
		oerr := StartOrchestrator(cfg)
		if oerr != nil {
			panic(log.Critical(oerr))
		}
	}()

	if cfg.GetBool("container.enableEthernetInspector") {
		nfqNum := cfg.GetInt("container.ethernetNFQNumber")
		if nfqNum <= 0 {
			return fmt.Errorf("strange container.ethernetNFQNumber: %d", nfqNum)
		}
		log.Debugf("Configuring NFQUEUE %d for container %s", nfqNum, c.ID)
		err := SetupNFQUEUE(c, nfqNum, false, false)
		if err != nil {
			return err
		}
		log.Debugf("Starting Ethernet Inspector")
		go func() {
			ierr := ServeEthernetInspector(c, nfqNum)
			if ierr != nil {
				panic(log.Critical(ierr))
			}
		}()
	}

	if cfg.GetBool("container.enableProcInspector") {
		watchInterval := cfg.GetDuration("container.procWatchInterval")
		if watchInterval <= 0 {
			return fmt.Errorf("strange container.procWatchInterval: %s", watchInterval)
		}
		log.Debugf("Starting Process Inspector")
		go func() {
			ierr := ServeProcInspector(c, watchInterval)
			if ierr != nil {
				panic(log.Critical(ierr))
			}
		}()
	}

	return nil
}

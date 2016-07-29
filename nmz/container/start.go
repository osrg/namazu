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
	"io/ioutil"
	"strings"

	log "github.com/cihub/seelog"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/osrg/namazu/nmz/util/config"
)

func StartNamazuRoutinesPre(dockerOpt *docker.CreateContainerOptions, cfg config.Config) (*docker.CreateContainerOptions, error) {
	log.Debugf("Starting Orchestrator")
	go func() {
		oerr := StartOrchestrator(cfg)
		if oerr != nil {
			panic(log.Critical(oerr))
		}
	}()

	var newBinds []string
	for _, bind := range dockerOpt.HostConfig.Binds {
		split := strings.Split(bind, ":")
		if len(split) != 2 {
			return dockerOpt, fmt.Errorf("bind is expected to be <foo>:<bar>, got %s", bind)
		}
		bindSrc, bindDst := split[0], split[1]
		mountpoint, err := ioutil.TempDir("", "nmz-container-fs-inspector")
		if err != nil {
			return dockerOpt, err
		}
		if cfg.GetBool("container.enableFSInspector") {
			log.Debugf("Starting Filesystem Inspector for %s (on %s)", bindSrc, mountpoint)
			log.Warnf("Please run `fusermount -i %s` manually on exit", mountpoint)
			go func() {
				ierr := ServeFSInspector(bindSrc, mountpoint)
				if ierr != nil {
					panic(log.Critical(ierr))
				}
			}()
		}
		newBinds = append(newBinds, fmt.Sprintf("%s:%s", mountpoint, bindDst))
	}
	dockerOpt.HostConfig.Binds = newBinds
	return dockerOpt, nil
}

func StartNamazuRoutinesPost(c *docker.Container, cfg config.Config) error {
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

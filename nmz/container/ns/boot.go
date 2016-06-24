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

package ns

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	dockerpty "github.com/fgrehm/go-dockerpty"
	docker "github.com/fsouza/go-dockerclient"
)

func Boot(client *docker.Client, opt *docker.CreateContainerOptions, detach bool,
	exitCh chan error) (*docker.Container, error) {
	log.Debugf("Creating container for image %s", opt.Config.Image)
	container, err := client.CreateContainer(*opt)
	if err != nil {
		return container, err
	}

	log.Debugf("Starting container %s", container.ID)
	go func() {
		if detach {
			err = client.StartContainer(container.ID, opt.HostConfig)
			if err == nil {
				_, err = client.WaitContainer(container.ID)
			}
			exitCh <- err
		} else {
			exitCh <- dockerpty.Start(client, container, opt.HostConfig)
		}
	}()

	trial := 0
	for {
		container, err = client.InspectContainer(container.ID)
		if container.State.StartedAt.Unix() > 0 {
			break
		}
		if trial > 30 {
			return container, fmt.Errorf("container %s seems not started. state=%#v", container.ID, container.State)
		}
		trial += 1
		time.Sleep(time.Duration(trial*100) * time.Millisecond)
	}
	log.Debugf("container state=%#v", container.State)
	return container, nil
}

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
	"runtime"

	log "github.com/cihub/seelog"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/vishvananda/netns"
)

var origNs netns.NsHandle
var newNs netns.NsHandle

func EnterDockerNetNs(container *docker.Container) error {
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

func LeaveNetNs() {
	netns.Set(origNs)
	runtime.UnlockOSThread()
	origNs.Close()
	newNs.Close()
}

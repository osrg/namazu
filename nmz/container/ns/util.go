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
	"os"
	"strings"

	log "github.com/cihub/seelog"
	docker "github.com/fsouza/go-dockerclient"
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

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

package run

import (
	"fmt"
	log "github.com/cihub/seelog"
	flag "github.com/docker/docker/pkg/mflag"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/osrg/earthquake/earthquake-container/container"
	"github.com/osrg/earthquake/earthquake-container/core"
	"github.com/osrg/earthquake/earthquake/util/config"
	"os"
)

func prepare(args []string) (dockerOpt *docker.CreateContainerOptions, removeOnExit bool, eqCfg config.Config, err error) {
	if len(args) < 3 {
		// FIXME
		err = fmt.Errorf("bad argument: %s", args)
		return
	}
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	dockerOpt, err = parseRun(flagSet, args[1:])
	if err != nil {
		return
	}
	removeOnExit = flagSet.IsSet("-rm")

	eqCfgPath := flagSet.Lookup("-eq-config").Value.String()
	eqCfg, err = newConfig(eqCfgPath)
	if err != nil {
		err = fmt.Errorf("bad config: %s", err)
		return
	}
	log.Debugf("Earthquake Config=%s", eqCfg)

	if err = checkPrerequisite(eqCfg); err != nil {
		err = fmt.Errorf("prerequisite error: %s", err)
	}
	return
}

func Run(args []string) int {
	dockerOpt, removeOnExit, eqCfg, err := prepare(args)
	if err != nil {
		// do not panic here
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}

	client, err := container.NewDockerClient()
	if err != nil {
		panic(log.Critical(err))
	}

	containerExitStatusChan := make(chan error)
	c, err := container.Boot(client, dockerOpt, containerExitStatusChan)
	if err == docker.ErrNoSuchImage {
		log.Critical(err)
		// TODO: pull the image automatically
		log.Infof("You need to run `docker pull %s`", dockerOpt.Config.Image)
		return 1
	} else if err != nil {
		panic(log.Critical(err))
	}
	if removeOnExit {
		defer container.Remove(client, c)
	}

	err = core.StartEarthquakeRoutines(c, eqCfg)
	if err != nil {
		panic(log.Critical(err))
	}

	err = <-containerExitStatusChan
	if err != nil {
		// do not panic here
		log.Error(err)
	}
	log.Debugf("Exiting..")
	// TODO: propagate err
	return 0
}

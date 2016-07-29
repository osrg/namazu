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
	"os"

	log "github.com/cihub/seelog"
	flag "github.com/docker/docker/pkg/mflag"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/osrg/namazu/nmz/container"
	"github.com/osrg/namazu/nmz/container/ns"
	"github.com/osrg/namazu/nmz/util/config"
)

func prepare(args []string) (dockerOpt *docker.CreateContainerOptions, removeOnExit bool, detach bool, nmzCfg config.Config, err error) {
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
	detach = flagSet.IsSet("d") || flag.IsSet("-detach")

	nmzCfgPath := flagSet.Lookup("-nmz-autopilot").Value.String()
	nmzCfg, err = newConfig(nmzCfgPath)
	if err != nil {
		err = fmt.Errorf("bad config: %s", err)
		return
	}
	log.Debugf("Namazu Config=%s", nmzCfg)

	if err = checkPrerequisite(nmzCfg); err != nil {
		err = fmt.Errorf("prerequisite error: %s", err)
	}
	return
}

func help() string {
	// FIXME: why not use the strings in runflag.go?
	s := `Usage: namazu container run [OPTIONS] IMAGE COMMAND

Run a command in a new Namazu Container

Docker-compatible options:
  -d, --detach                    Run container in background and print container ID (nmz itself runs in foreground)
  -i, --interactive               Keep STDIN open even if not attached
  -p                              Publish a container's port(s) to the host
  --name                          Assign a name to the container
  --rm                            Automatically remove the container when it exits
  -t, --tty                       Allocate a pseudo-TTY
  -v, --volume=[]                 Bind mount a volume
  --volumes-from                  Mount volumes from the specified container(s)
  --privileged                    Give extended privileges to this container

Namazu-specific options:
  -nmz-autopilot                      Namazu configuration file

NOTE: Unlike docker, COMMAND is mandatory at the moment.
`
	return s
}

func Run(args []string) int {
	dockerOpt, removeOnExit, detach, nmzCfg, err := prepare(args)
	if err != nil {
		// do not panic here
		fmt.Fprintf(os.Stderr, "%s\n", err)
		fmt.Fprintf(os.Stderr, "\n%s\n", help())
		return 1
	}

	dockerOpt, err = container.StartNamazuRoutinesPre(dockerOpt, nmzCfg)
	if err != nil {
		panic(log.Critical(err))
	}

	client, err := ns.NewDockerClient()
	if err != nil {
		panic(log.Critical(err))
	}

	containerExitStatusChan := make(chan error)
	c, err := ns.Boot(client, dockerOpt, detach, containerExitStatusChan)
	if err == docker.ErrNoSuchImage {
		log.Critical(err)
		// TODO: pull the image automatically
		log.Infof("You need to run `docker pull %s`", dockerOpt.Config.Image)
		return 1
	} else if err != nil {
		panic(log.Critical(err))
	}
	if removeOnExit {
		defer ns.Remove(client, c)
	}

	if detach {
		fmt.Printf("%s\n", c.ID)
		log.Info("Namazu container is running the container in background, but Namazu itself keeps running in foreground.")
	}

	err = container.StartNamazuRoutinesPost(c, nmzCfg)
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

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

package cli

import (
	"fmt"
	log "github.com/cihub/seelog"
	flag "github.com/docker/docker/pkg/mflag"
	dockerpty "github.com/fgrehm/go-dockerpty"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/osrg/earthquake/earthquake-container/container"
	"os"
	"time"
)

func parseRun(cmd *flag.FlagSet, args []string) (*docker.CreateContainerOptions, error) {
	var (
		flStdin  = cmd.Bool([]string{"i", "-interactive"}, false, "Keep STDIN open even if not attached")
		flTty    = cmd.Bool([]string{"t", "-tty"}, false, "Allocate a pseudo-TTY")
		flDetach = cmd.Bool([]string{"d", "-detach"}, false, "[NOT SUPPORTED] Run container in background and print container ID")
		flName   = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		// the caller should handle "-rm" with cmd.IsSet()
		_ = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")
		// the caller should handle "-eq-config"
		flEqConfig = cmd.String([]string{"-eq-config"}, "", "Earthquake configuration file")
	)
	if err := cmd.Parse(args); err != nil {
		return nil, err
	}
	if !*flStdin {
		return nil, fmt.Errorf("-interactive is expected.")
	}
	if !*flTty {
		return nil, fmt.Errorf("-tty is expected.")
	}
	if *flDetach {
		return nil, fmt.Errorf("Currently, -detach is not supported.")
	}
	if *flEqConfig == "" {
		return nil, fmt.Errorf("-eq-config is expected.")
	}

	parsedArgs := cmd.Args()
	if len(parsedArgs) < 2 {
		return nil, fmt.Errorf("requires a minimum of 2 arguments")
	}
	image := parsedArgs[0]
	execCmd := parsedArgs[1:]

	dockerOpt := docker.CreateContainerOptions{
		Name: *flName,
		Config: &docker.Config{
			Image:        image,
			Cmd:          execCmd,
			OpenStdin:    *flStdin,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          *flTty,
		},
		HostConfig: &docker.HostConfig{},
	}
	return &dockerOpt, nil
}

func bootContainer(client *docker.Client, opt *docker.CreateContainerOptions,
	exitCh chan error) (*docker.Container, error) {
	log.Debugf("Creating container for image %s", opt.Config.Image)
	container, err := client.CreateContainer(*opt)
	if err != nil {
		return container, err
	}

	log.Debugf("Starting container %s", container.ID)
	go func() {
		exitCh <- dockerpty.Start(client, container, opt.HostConfig)
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

func removeContainer(client *docker.Client, container *docker.Container) error {
	log.Debugf("Removing container %s", container.ID)
	err := client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    container.ID,
		Force: true,
	})
	log.Debugf("Removed container %s", container.ID)
	if err != nil {
		log.Error(err)
	}
	return err
}

func startRoutines(container *docker.Container, nfqueueNum int, configPath string) error {
	log.Debugf("Configuring NFQUEUE %d for container %s", nfqueueNum, container.ID)
	err := SetupNFQUEUE(container, nfqueueNum, false, false)
	if err != nil {
		return err
	}

	// TODO: refactor
	log.Debugf("Starting Orchestrator")
	go func() {
		oerr := StartOrchestrator(configPath)
		if oerr != nil {
			panic(log.Critical(oerr))
		}
	}()

	log.Debugf("Starting Inspector")
	go func() {
		ierr := StartEthernetInspector(container, nfqueueNum)
		if ierr != nil {
			panic(log.Critical(ierr))
		}
	}()

	return nil
}

// FIXME: too long function scope
func run(args []string) int {
	if len(args) < 3 {
		// FIXME
		fmt.Fprintf(os.Stderr, "bad argument: %s\n", args)
		return 1
	}
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	dockerOpt, err := parseRun(flagSet, args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	removeOnExit := flagSet.IsSet("-rm")

	nfqueueNum := 42 // FIXME
	configPath := flagSet.Lookup("-eq-config").Value.String()

	if err = checkPrerequisite(); err != nil {
		fmt.Fprintf(os.Stderr, "prerequisite error: %s\n", err)
		return 1
	}

	client, err := NewDockerClient()
	if err != nil {
		panic(err)
	}

	exited := make(chan error)
	container, err := bootContainer(client, dockerOpt, exited)
	if err == docker.ErrNoSuchImage {
		log.Critical(err)
		// TODO: pull the image automatically
		log.Infof("You need to run `docker pull %s`", dockerOpt.Config.Image)
		return 1
	} else if err != nil {
		panic(err)
	}
	if removeOnExit {
		defer removeContainer(client, container)
	}

	err = startRoutines(container, nfqueueNum, configPath)
	if err != nil {
		panic(err)
	}

	err = <-exited
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Exiting..")

	return 0
}

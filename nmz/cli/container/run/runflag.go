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
	// FIXME: we should not rely on internal docker packages..
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/go-connections/nat"
	docker "github.com/fsouza/go-dockerclient"
)

// TODO: we should support more options..
func parseRun(cmd *flag.FlagSet, args []string) (*docker.CreateContainerOptions, error) {
	var (
		attachStdin  = true
		attachStdout = true
		attachStderr = true
		stdinOnce    = true

		flVolumes     = opts.NewListOpts(nil)
		flVolumesFrom = opts.NewListOpts(nil)
		flPublish     = opts.NewListOpts(nil)

		flStdin      = cmd.Bool([]string{"i", "-interactive"}, false, "Keep STDIN open even if not attached")
		flTty        = cmd.Bool([]string{"t", "-tty"}, false, "Allocate a pseudo-TTY")
		flPrivileged = cmd.Bool([]string{"-privileged"}, false, "Give extended privileges to this container")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Run container in background and print container ID")
		flAutoRemove = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")

		flName    = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		flNetMode = cmd.String([]string{"-net"}, "default", "Connect a container to a network")

		// the caller should handle "-nmz-autopilot"
		_ = cmd.String([]string{"-nmz-autopilot"}, "", "Namazu configuration file")
	)
	cmd.Var(&flVolumes, []string{"v", "-volume"}, "Bind mount a volume")
	cmd.Var(&flPublish, []string{"p"}, "Publish a container's port(s) to the host")
	cmd.Var(&flVolumesFrom, []string{"-volumes-from"}, "Mount volumes from the specified container(s)")

	if err := cmd.Parse(args); err != nil {
		return nil, err
	}

	parsedArgs := cmd.Args()

	if !*flDetach {
		if !*flStdin {
			return nil, fmt.Errorf("--interactive is expected in attach mode.")
		}
		if !*flTty {
			return nil, fmt.Errorf("--tty is expected in attach mode.")
		}
		if len(parsedArgs) < 2 {
			return nil, fmt.Errorf("requires a minimum of 2 arguments in attach mode.")
		}
	} else {
		if *flAutoRemove {
			return nil, fmt.Errorf("--rm is not availible in detach mode.")
		}
		attachStdin = false
		attachStdout = false
		attachStderr = false
		stdinOnce = false
	}

	image := parsedArgs[0]
	execCmd := parsedArgs[1:]

	// FIXME: we should implement this parse function ourselves..
	_, portMap, err := nat.ParsePortSpecs(flPublish.GetAll())
	if err != nil {
		return nil, err
	}

	// Transform nat.PortMap to docker.PortMap(map[string][]PortBindings).
	portBindings := make(map[docker.Port][]docker.PortBinding)
	for port, bindings := range portMap {
		for _, binding := range bindings {
			// nat.Port string
			bindingSlice, exists := portBindings[docker.Port(port)]
			if !exists {
				bindingSlice = []docker.PortBinding{}
			}

			portBindings[docker.Port(port)] = append(bindingSlice, docker.PortBinding{
				HostIP:   binding.HostIP,
				HostPort: binding.HostPort,
			})
		}
	}

	dockerOpt := docker.CreateContainerOptions{
		Name: *flName,
		Config: &docker.Config{
			Image:        image,
			Cmd:          execCmd,
			OpenStdin:    *flStdin,
			StdinOnce:    stdinOnce,
			AttachStdin:  attachStdin,
			AttachStdout: attachStdout,
			AttachStderr: attachStderr,
			Tty:          *flTty,
		},
		HostConfig: &docker.HostConfig{
			Binds:        flVolumes.GetAllOrEmpty(),
			Privileged:   *flPrivileged,
			PortBindings: portBindings,
			NetworkMode:  *flNetMode,
			VolumesFrom:  flVolumesFrom.GetAllOrEmpty(),
		},
	}
	return &dockerOpt, nil
}

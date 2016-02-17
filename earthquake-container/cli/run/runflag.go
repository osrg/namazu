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

	flag "github.com/docker/docker/pkg/mflag"
	docker "github.com/fsouza/go-dockerclient"
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
		_ = cmd.String([]string{"-eq-config"}, "", "Earthquake configuration file")
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

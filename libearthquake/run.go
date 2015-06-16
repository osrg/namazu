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

package main

import (
	// "encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	. "./equtils"

	"github.com/mitchellh/cli"
)

type runConfig struct {
	runScript string
}

func parseRunConfig(jsonPath string) (*runConfig, error) {
	jsonBuf, rerr := WholeRead(jsonPath)
	if rerr != nil {
		return nil, rerr
	}

	var root map[string]interface{}
	err := json.Unmarshal(jsonBuf, &root)
	if err != nil {
		return nil, err
	}

	runScript := root["run"].(string)

	return &runConfig{
		runScript: runScript,
	}, nil
}

func run(args []string) {
	if len(args) != 1 {
		fmt.Printf("specify <storage dir path>\n")
		os.Exit(1)
	}

	storage := args[0]
	confPath := storage + "/" + storageConfigPath

	conf, err := parseRunConfig(confPath)
	if err != nil {
		fmt.Printf("failed to parse config file %s: %s\n", confPath, err)
		os.Exit(1)
	}

	info := readSearchModeDir(storage)
	fmt.Printf("NrCollectedTraces: %d\n", info.NrCollectedTraces)

	materialsDir := storage + "/" + storageMaterialsPath
	runScriptPath := materialsDir + "/" + conf.runScript
	runCmd := exec.Command("sh", "-c", runScriptPath)

	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	runCmd.Env = append(runCmd.Env, "MATERIALS_DIR=" + materialsDir)

	rerr := runCmd.Run()
	if rerr != nil {
		fmt.Printf("failed to execute run script %s: %s\n", runScriptPath, rerr)
		os.Exit(1)
	}
}

type runCmd struct {
}

func (cmd runCmd) Help() string {
	return "run help (todo)"
}

func (cmd runCmd) Run(args []string) int {
	run(args)
	return 0
}

func (cmd runCmd) Synopsis() string {
	return "run subcommand"
}

func runCommandFactory() (cli.Command, error) {
	return runCmd{}, nil
}

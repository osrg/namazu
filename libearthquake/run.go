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
	"time"

	. "./equtils"

	"./historystorage"
	"./searchpolicy"

	"github.com/mitchellh/cli"
)

type runConfig struct {
	runScript      string
	cleanScript    string
	validateScript string

	searchPolicy string
	storageType  string

	searchPolicyParam map[string]interface{}
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

	runScript := ""
	if _, ok := root["run"]; ok {
		runScript = root["run"].(string)
	} else {
		fmt.Printf("required field \"run\" is missing\n")
		os.Exit(1) // TODO: construct suitable error
		return nil, nil
	}

	cleanScript := ""
	if _, ok := root["clean"]; ok {
		cleanScript = root["clean"].(string)
	}

	validateScript := ""
	if _, ok := root["validate"]; ok {
		validateScript = root["validate"].(string)
	}

	searchPolicy := "dumb"
	var searchPolicyParam map[string]interface{}
	if _, ok := root["searchPolicy"]; ok {
		searchPolicy = root["searchPolicy"].(string)

		if _, ok := root["searchPolicyParam"]; ok {
			searchPolicyParam = root["searchPolicyParam"].(map[string]interface{})
		}
	}

	storageType := "naive"
	if _, ok := root["storageType"]; ok {
		storageType = root["storageType"].(string)
	}

	return &runConfig{
		runScript:      runScript,
		cleanScript:    cleanScript,
		validateScript: validateScript,

		searchPolicy: searchPolicy,
		storageType:  storageType,

		searchPolicyParam: searchPolicyParam,
	}, nil
}

func createCmd(scriptPath, workingDirPath, materialsDirPath string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", scriptPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(cmd.Env, "WORKING_DIR="+workingDirPath)
	cmd.Env = append(cmd.Env, "MATERIALS_DIR="+materialsDirPath)

	return cmd
}

func run(args []string) {
	if len(args) != 1 {
		fmt.Printf("specify <storage dir path>\n")
		os.Exit(1)
	}

	storagePath := args[0]
	confPath := storagePath + "/" + storageConfigPath

	conf, err := parseRunConfig(confPath)
	if err != nil {
		fmt.Printf("failed to parse config file %s: %s\n", confPath, err)
		os.Exit(1)
	}

	storage := historystorage.New(conf.storageType, storagePath)
	storage.Init()

	policy := searchpolicy.CreatePolicy(conf.searchPolicy)
	if policy == nil {
		fmt.Printf("invalid policy name: %s", conf.searchPolicy)
		os.Exit(1)
	}
	policy.Init(storage, conf.searchPolicyParam)

	nextDir := storage.CreateNewWorkingDir()
	InitLog(nextDir + "/earthquake.log")

	end := make(chan interface{})
	newTraceCh := make(chan *SingleTrace)

	go searchModeNoInitiation(nextDir, policy, end, newTraceCh)

	materialsDir := storagePath + "/" + storageMaterialsPath
	runScriptPath := materialsDir + "/" + conf.runScript

	cleanScriptPath := ""
	if conf.cleanScript != "" {
		cleanScriptPath = materialsDir + "/" + conf.cleanScript
	}

	validateScriptPath := ""
	if conf.validateScript != "" {
		validateScriptPath = materialsDir + "/" + conf.validateScript
	}

	runCmd := createCmd(runScriptPath, nextDir, materialsDir)

	startTime := time.Now()

	rerr := runCmd.Run()
	if rerr != nil {
		fmt.Printf("failed to execute run script %s: %s\n", runScriptPath, rerr)
		os.Exit(1)
	}

	end <- true
	newTrace := <-newTraceCh

	endTime := time.Now()
	requiredTime := endTime.Sub(startTime)

	storage.RecordNewTrace(newTrace)

	if validateScriptPath != "" {
		validateCmd := createCmd(validateScriptPath, nextDir, materialsDir)

		rerr = validateCmd.Run()
		if rerr != nil {
			fmt.Printf("validation failed: %s\n", rerr)
			// TODO: detailed check of error
			// e.g. handle a case like permission denied, noent, etc
			storage.RecordResult(false, requiredTime)
		} else {
			fmt.Printf("validation succeed\n")
			storage.RecordResult(true, requiredTime)
		}
	}

	if cleanScriptPath != "" {
		cleanCmd := createCmd(cleanScriptPath, nextDir, materialsDir)

		rerr = cleanCmd.Run()
		if rerr != nil {
			fmt.Printf("failed to execute clean script %s: %s\n", cleanScriptPath, rerr)
			os.Exit(1)
		}
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

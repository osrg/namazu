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
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	. "./equtils"

	"./explorepolicy"
	"./historystorage"

	"github.com/mitchellh/cli"
)

func init() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
	}

	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
	}
}

func createCmd(scriptPath, workingDirPath, materialsDirPath string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", scriptPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ() // this line is needed to extend current envs
	if workingDirPath != "" { // can be empty for "init"
		cmd.Env = append(cmd.Env, "EQ_WORKING_DIR="+workingDirPath)
	}
	cmd.Env = append(cmd.Env, "EQ_MATERIALS_DIR="+materialsDirPath)

	return cmd
}

func run(args []string) {
	if len(args) != 1 {
		fmt.Printf("specify <storage dir path>\n")
		os.Exit(1)
	}

	storagePath := args[0]
	confPath := storagePath + "/" + historystorage.StorageConfigPath

	config, err := ParseConfigFile(confPath)
	if err != nil {
		fmt.Printf("failed to parse config file %s: %s\n", confPath, err)
		os.Exit(1)
	}

	storage := historystorage.New(config.GetString("storageType"), storagePath)
	storage.Init()

	policy := explorepolicy.CreatePolicy(config.GetString("explorePolicy"))
	if policy == nil {
		fmt.Printf("invalid policy name: %s", config.GetString("explorePolicy"))
		os.Exit(1)
	}
	policy.Init(storage, config.GetStringMap("explorePolicyParam"))

	nextDir := storage.CreateNewWorkingDir()
	InitLog(nextDir + "/earthquake.log")
	AddLogTee(os.Stdout)

	end := make(chan interface{})
	newTraceCh := make(chan *SingleTrace)

	go orchestrate(end, policy, newTraceCh)

	materialsDir := storagePath + "/" + storageMaterialsPath
	runScriptPath := materialsDir + "/" + config.GetString("run")

	cleanScriptPath := ""
	if config.GetString("clean") != "" {
		cleanScriptPath = materialsDir + "/" + config.GetString("clean")
	}

	validateScriptPath := ""
	if config.GetString("validate") != "" {
		validateScriptPath = materialsDir + "/" + config.GetString("validate")
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

	succeed := true

	if validateScriptPath != "" {
		validateCmd := createCmd(validateScriptPath, nextDir, materialsDir)

		rerr = validateCmd.Run()
		if rerr != nil {
			fmt.Printf("validation failed: %s\n", rerr)
			// TODO: detailed check of error
			// e.g. handle a case like permission denied, noent, etc
			storage.RecordResult(false, requiredTime)
			succeed = false
		} else {
			fmt.Printf("validation succeed\n")
			storage.RecordResult(true, requiredTime)
		}
	}

	if succeed && cleanScriptPath != "" {
		cleanCmd := createCmd(cleanScriptPath, nextDir, materialsDir)

		rerr = cleanCmd.Run()
		if rerr != nil {
			fmt.Printf("failed to execute clean script %s: %s\n", cleanScriptPath, rerr)
			os.Exit(1)
		}
	}

	if !succeed {
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

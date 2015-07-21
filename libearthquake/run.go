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
	"time"

	. "./equtils"

	"./historystorage"
	"./explorepolicy"

	"github.com/mitchellh/cli"
)



func createCmd(scriptPath, workingDirPath, materialsDirPath string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", scriptPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ() // this line is needed to extend current envs
	cmd.Env = append(cmd.Env, "EQ_WORKING_DIR="+workingDirPath)
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

	vcfg, err := ParseConfigFile(confPath)
	if err != nil {
		fmt.Printf("failed to parse config file %s: %s\n", confPath, err)
		os.Exit(1)
	}

	storage := historystorage.New(vcfg.GetString("storageType"), storagePath)
	storage.Init()

	policy := explorepolicy.CreatePolicy(vcfg.GetString("explorePolicy"))
	if policy == nil {
		fmt.Printf("invalid policy name: %s", vcfg.GetString("explorePolicy"))
		os.Exit(1)
	}
	policy.Init(storage, vcfg.GetStringMap("explorePolicyParam"))

	nextDir := storage.CreateNewWorkingDir()
	InitLog(nextDir + "/earthquake.log")
	AddLogTee(os.Stdout)

	end := make(chan interface{})
	newTraceCh := make(chan *SingleTrace)

	go orchestrate(end, policy, newTraceCh)

	materialsDir := storagePath + "/" + storageMaterialsPath
	runScriptPath := materialsDir + "/" + vcfg.GetString("run")

	cleanScriptPath := ""
	if vcfg.GetString("clean") != "" {
		cleanScriptPath = materialsDir + "/" + vcfg.GetString("clean")
	}

	validateScriptPath := ""
	if vcfg.GetString("validate") != "" {
		validateScriptPath = materialsDir + "/" + vcfg.GetString("validate")
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

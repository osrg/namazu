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
	"path"
	"syscall"
	"time"

	"fmt"
	"os/exec"

	log "github.com/cihub/seelog"
	mcli "github.com/mitchellh/cli"
	"github.com/osrg/namazu/nmz/explorepolicy"
	"github.com/osrg/namazu/nmz/historystorage"
	"github.com/osrg/namazu/nmz/orchestrator"
	"github.com/osrg/namazu/nmz/util/cmd"
	"github.com/osrg/namazu/nmz/util/config"
	logutil "github.com/osrg/namazu/nmz/util/log"
)

func setRlimit() error {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("could not get Rlimit: %s", err)
	}

	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("could not set Rlimit: %s", err)
	}
	return nil
}

type runner struct {
	storageDirPath   string
	workingDirPath   string
	materialsDirPath string
	config           config.Config
	storage          historystorage.HistoryStorage
	policy           explorepolicy.ExplorePolicy
	runCmd           *exec.Cmd
	validateCmd      *exec.Cmd
	cleanCmd         *exec.Cmd
}

// depends on this.storageDirPath
func (this *runner) initConfig() error {
	var err error
	confPath := path.Join(this.storageDirPath, historystorage.StorageTOMLConfigPath)
	this.config, err = config.NewFromFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to parse config file %s: %s", confPath, err)
	}
	return nil
}

// depends on initConfig()
func (this *runner) initStorage() error {
	var err error
	this.storage, err = historystorage.New(this.config.GetString("storageType"), this.storageDirPath)
	if err != nil {
		return err
	}
	this.storage.Init()
	this.workingDirPath = this.storage.CreateNewWorkingDir()
	this.materialsDirPath = path.Join(this.storageDirPath, storageMaterialsPath)
	return nil
}

// depends on initStorage()
func (this *runner) initGlobalLogger() error {
	logutil.InitLog(path.Join(this.workingDirPath, "earthquake.log"), logutil.Debug)
	return nil
}

// depends on initStorage()
func (this *runner) initGlobalCmdFactory() error {
	cmd.DefaultFactory.SetWorkingDir(this.workingDirPath)
	cmd.DefaultFactory.SetMaterialsDir(this.materialsDirPath)
	return nil
}

// depends on initStorage(), initConfig(), initGlobalCmdFactory()
func (this *runner) initCmd() error {
	factory := cmd.DefaultFactory
	runPath := this.config.GetString("run")
	if runPath == "" {
		return fmt.Errorf("\"run\" is not set in the config")
	}
	this.runCmd = factory.CreateCmd(path.Join(this.materialsDirPath, runPath))

	validatePath := this.config.GetString("validate")
	if validatePath != "" {
		this.validateCmd = factory.CreateCmd(path.Join(this.materialsDirPath, validatePath))
	}

	cleanPath := this.config.GetString("clean")
	if cleanPath != "" {
		this.cleanCmd = factory.CreateCmd(path.Join(this.materialsDirPath, cleanPath))
	}
	return nil
}

// depends on initConfig(), initStorage()
func (this *runner) initPolicy() error {
	var err error
	this.policy, err = explorepolicy.CreatePolicy(this.config.GetString("explorePolicy"))
	if err != nil {
		return err
	}
	if err = this.policy.LoadConfig(this.config); err != nil {
		return err
	}
	if err = this.policy.SetHistoryStorage(this.storage); err != nil {
		return err
	}
	return nil
}

func newRunner(storageDirPath string) (*runner, error) {
	var err error
	r := &runner{
		storageDirPath: storageDirPath,
	}
	if err = r.initConfig(); err != nil {
		return nil, err
	}
	if err = r.initStorage(); err != nil {
		return nil, err
	}
	if err = r.initGlobalLogger(); err != nil {
		return nil, err
	}
	if err = r.initGlobalCmdFactory(); err != nil {
		return nil, err
	}
	if err = r.initCmd(); err != nil {
		return nil, err
	}
	if err = r.initPolicy(); err != nil {
		return nil, err
	}
	return r, nil
}

func runCommand(x *exec.Cmd) error {
	log.Infof("Starting %s %s", x.Path, x.Args)
	err := x.Run()
	log.Infof("Finished %s %s", x.Path, x.Args)
	return err
}

func run(args []string) int {
	// Parse args
	if len(args) != 1 {
		fmt.Printf("specify <storage dir path>\n")
		return 1 // panic() looks ugly for such an expected error
	}
	storagePath := args[0]

	// Initialize runner
	runner, err := newRunner(storagePath)
	if err != nil {
		panic(log.Critical(err))
	}

	// Set rlimit
	if err = setRlimit(); err != nil {
		// this is not a critical error
		log.Warn(err)
	}

	// Start orchestrator
	log.Infof("Starting Orchestrator with exploration policy \"%s\"", runner.policy.Name())
	orchestrator := orchestrator.NewOrchestrator(runner.config, runner.policy, true)
	go orchestrator.Start()
	log.Infof("Started Orchestrator")

	// Run
	startTime := time.Now()
	err = runCommand(runner.runCmd)
	if err != nil {
		log.Criticalf("failed to execute run script: %s\n", err)
		return 1
	}

	// Stop orchestrator
	log.Infof("Shutting down Orchestrator")
	trace := orchestrator.Shutdown()
	endTime := time.Now()
	requiredTime := endTime.Sub(startTime)
	log.Infof("Shut down Orchestrator (got %d actions, took %s)",
		len(trace.ActionSequence), requiredTime)

	// Validate
	successful := true
	if runner.validateCmd != nil {
		if err = runCommand(runner.validateCmd); err != nil {
			log.Infof("Validation failed: %s", err)
			// TODO: detailed check of error
			// e.g. handle a case like permission denied, noent, etc
			successful = false
		} else {
			log.Infof("Validation succeeded")
		}
	} else {
		log.Warn("No validation script provided")
	}

	// Record
	runner.storage.RecordNewTrace(trace)
	runner.storage.RecordResult(successful, requiredTime)
	runner.storage.Close()

	// Clean
	if successful || !runner.config.GetBool("notCleanIfValidationFail") {
		if runner.cleanCmd != nil {
			if err = runCommand(runner.cleanCmd); err != nil {
				log.Criticalf("failed to execute clean script: %s", err)
				return 1
			}
		}
	}

	return 0
}

type runCmd struct {
}

func (cmd runCmd) Help() string {
	s := `
The run command starts the orchestrator and run an experiment with the workspace.
Before running this command, you have to initialize the workspace using the init command.

Typical usage:
     $ earthquake init --force config.toml materials /tmp/x
     $ for f in $(seq 1 100);do earthquake run /tmp/x; done
     $ earthquake tools summary /tmp/x

You have to prepare config.toml and the materials directory before running the init command.
Please also refer to the examples included in the github repository: https://github.com/osrg/namazu/tree/master/example

NOTE: "earthquake run" is different from "earthquake container run".
`
	return s
}

func (cmd runCmd) Run(args []string) int {
	log.Warn("`earthquake run` is different from `earthquake container run` (Docker-like CLI).")
	return run(args)
}

func (cmd runCmd) Synopsis() string {
	return "[Expert] Run an experiment with the initialized workspace"
}

func runCommandFactory() (mcli.Command, error) {
	return runCmd{}, nil
}

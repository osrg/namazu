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

// Package cmd provides utilities for running shell commands
package cmd

import (
	"os"
	"os/exec"

	log "github.com/cihub/seelog"
)

type CmdFactory struct {
	workingDir   string
	materialsDir string
}

// set NMZ_WORKING_DIR
func (this *CmdFactory) SetWorkingDir(s string) {
	this.workingDir = s
}

func (this *CmdFactory) GetWorkingDir() string {
	return this.workingDir
}

// set NMZ_MATERIALS_DIR
func (this *CmdFactory) SetMaterialsDir(s string) {
	this.materialsDir = s
}

func (this *CmdFactory) GetMaterialsDir() string {
	return this.materialsDir
}

func (this *CmdFactory) CreateCmd(scriptPath string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", scriptPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// this line is needed to extend current envs
	cmd.Env = os.Environ()

	// workinDir can be empty for `nmz init`
	if this.workingDir != "" {
		cmd.Env = append(cmd.Env, "NMZ_WORKING_DIR="+this.workingDir)
	}

	if this.materialsDir != "" {
		cmd.Env = append(cmd.Env, "NMZ_MATERIALS_DIR="+this.materialsDir)
	} else {
		log.Warnf("MaterialsDir is empty")
	}
	return cmd
}

func NewCmdFactory() *CmdFactory {
	return &CmdFactory{}
}

var DefaultFactory *CmdFactory

func init() {
	DefaultFactory = NewCmdFactory()
}

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
	"fmt"
	"os"

	// . "../equtils"

	"github.com/mitchellh/cli"
)

const (
	storageConfigPath string = "config.json"
)

func _init(args []string) {
	if len(args) != 2 {
		fmt.Printf("specify <config file path> <storage dir path>\n")
		os.Exit(1)
	}

	conf := args[0]
	storage := args[1]

	cfi, cerr := os.Stat(conf)
	if cerr != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", conf, cerr)
		os.Exit(1)
	}

	if !cfi.Mode().IsRegular() {
		fmt.Printf("config file (%s) must be a regular file\n", conf)
		os.Exit(1)
	}

	sfi, serr := os.Stat(storage)
	if serr != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", storage, serr)
		os.Exit(1)
	}

	if !sfi.Mode().IsDir() {
		fmt.Printf("storage directory (%s) must be a directory\n", storage)
		os.Exit(1)
	}

	dir, derr := os.Open(storage)
	if derr != nil {
		fmt.Printf("failed to open storage directory: %s (%s)\n", storage, derr)
		os.Exit(1)
	}

	fi, rderr := dir.Readdir(0)
	if rderr != nil {
		fmt.Printf("failed to read storage directory: %s (%s)\n", storage, rderr)
		os.Exit(1)
	}

	if len(fi) != 0 {
		fmt.Printf("directory for earthquake storage (%s) must be empty\n", storage)
		os.Exit(1)
	}

	lerr := os.Link(conf, storage + "/" + storageConfigPath)
	if lerr != nil {
		fmt.Printf("creating link of config file (%s) failed (%s)\n", conf, lerr)
		os.Exit(1)
	}

	fmt.Printf("ok\n")
}

type initCmd struct {
}

func (cmd initCmd) Help() string {
	return "init help (todo)"
}

func (cmd initCmd) Run(args []string) int {
	_init(args)
	return 0
}

func (cmd initCmd) Synopsis() string {
	return "init subcommand"
}

func initCommandFactory() (cli.Command, error) {
	return initCmd{}, nil
}

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

	. "./equtils"

	"./historystorage"
	"github.com/mitchellh/cli"
)

const (
	storageMaterialsPath string = "materials"
)

func recursiveHardLink(srcPath, dstPath string) {
	// TODO: write error to stderr with some logging library
	f, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("failed to open source path: %s (%s)\n",
			srcPath, err)
		os.Exit(1)
	}

	names, err := f.Readdirnames(0)
	if err != nil {
		fmt.Printf("failed to readdirnames: %s\n", err)
		os.Exit(1)
	}

	for _, name := range names {
		path := srcPath + "/" + name

		fi, err := os.Lstat(path)
		if err != nil {
			fmt.Printf("failed to stat (%s): %s", path, err)
			os.Exit(1)
		}

		if fi.Mode().IsDir() {
			dstDir := dstPath + "/" + name
			err := os.Mkdir(dstDir, 0777)
			if err != nil {
				fmt.Printf("failed to make directory %s: %s\n",
					dstDir, err)
				os.Exit(1)
			}
			recursiveHardLink(path, dstDir)
		} else {
			realPath := path
			if fi.Mode() & os.ModeSymlink != 0 {
				realPath, err = os.Readlink(path)
				if err != nil {
					fmt.Printf("could not read link %s", path)
					os.Exit(1)
				}
			}
			err := os.Link(realPath, dstPath+"/"+name)
			if err != nil {
				fmt.Printf("failed to link (src: %s(%s), dst: %s): %s\n",
					path, realPath, dstPath+"/"+name, err)
				os.Exit(1)
			}
		}
	}

}

func _init(args []string) {
	if len(args) != 3 {
		fmt.Printf("specify <config file path> <materials dir path> <storage dir path>\n")
		os.Exit(1)
	}

	conf := args[0]
	materials := args[1]
	storagePath := args[2]

	cfi, cerr := os.Stat(conf)
	if cerr != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", conf, cerr)
		os.Exit(1)
	}

	if !cfi.Mode().IsRegular() {
		fmt.Printf("config file (%s) must be a regular file\n", conf)
		os.Exit(1)
	}

	sfi, serr := os.Stat(storagePath)
	if serr != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", storagePath, serr)
		os.Exit(1)
	}

	if !sfi.Mode().IsDir() {
		fmt.Printf("storagePath directory (%s) must be a directory\n", storagePath)
		os.Exit(1)
	}

	dir, derr := os.Open(storagePath)
	if derr != nil {
		fmt.Printf("failed to open storagePath directory: %s (%s)\n", storagePath, derr)
		os.Exit(1)
	}

	fi, rderr := dir.Readdir(0)
	if rderr != nil {
		fmt.Printf("failed to read storagePath directory: %s (%s)\n", storagePath, rderr)
		os.Exit(1)
	}

	if len(fi) != 0 {
		fmt.Printf("directory for earthquake storagePath (%s) must be empty\n", storagePath)
		os.Exit(1)
	}

	vcfg, err := ParseConfigFile(conf)
	if err != nil {
		fmt.Printf("parsing config file (%s) failed: %s\n", conf, err)
		os.Exit(1)
	}

	lerr := os.Link(conf, storagePath+"/"+historystorage.StorageConfigPath)
	if lerr != nil {
		fmt.Printf("creating link of config file (%s) failed (%s)\n", conf, lerr)
		os.Exit(1)
	}

	materialDir := storagePath + "/" + storageMaterialsPath
	derr = os.Mkdir(materialDir, 0777)
	if derr != nil {
		fmt.Printf("creating a directory for materials (%s) failed (%s)\n",
			storagePath+"/"+storageMaterialsPath, derr)
		os.Exit(1)
		// TODO: cleaning conf file
	}

	recursiveHardLink(materials, materialDir)

	storage := historystorage.New(vcfg.GetString("storageType"), storagePath)
	storage.CreateStorage()

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

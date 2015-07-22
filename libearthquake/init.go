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
	"path"
	. "./equtils"

	"./historystorage"
	"github.com/mitchellh/cli"
	"flag"
)

const (
	storageMaterialsPath string = "materials"
)
type initFlags struct {
	Force bool
}

var (
	initFlagset = flag.NewFlagSet("init", flag.ExitOnError)
	_initFlags  = initFlags{}
)

func init() {
	initFlagset.BoolVar(&_initFlags.Force, "force", false, "forcibly (re-)create storage dir")
}

func recursiveHardLink(srcPath, dstPath string) error {
	// TODO: write error to stderr with some logging library
	f, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("failed to open source path: %s (%s)\n",
			srcPath, err)
		return err
	}

	names, err := f.Readdirnames(0)
	if err != nil {
		fmt.Printf("failed to readdirnames: %s\n", err)
		return err
	}

	for _, name := range names {
		p := path.Join(srcPath, name)

		fi, err := os.Lstat(p)
		if err != nil {
			fmt.Printf("failed to stat (%s): %s", p, err)
			return err
		}

		if fi.Mode().IsDir() {
			dstDir := path.Join(dstPath, name)
			err = os.Mkdir(dstDir, 0777)
			if err != nil {
				fmt.Printf("failed to make directory %s: %s\n",
					dstDir, err)
				return err
			}
			err = recursiveHardLink(p, dstDir)
			if err != nil {
				return err
			}
		} else {
			realPath := p
			if fi.Mode() & os.ModeSymlink != 0 {
				realPath, err = os.Readlink(p)
				if err != nil {
					fmt.Printf("could not read link %s", p)
					return err
				}
			}
			err = os.Link(realPath, path.Join(dstPath, name))
			if err != nil {
				fmt.Printf("failed to link (src: %s(%s), dst: %s): %s\n",
					p, realPath, path.Join(dstPath, name), err)
				return err
			}
		}
	}
	return nil
}

func _init(args []string) {
	if err := initFlagset.Parse(args); err != nil {
		fmt.Printf("%s", err.Error())
		os.Exit(1)
	}

	if initFlagset.NArg() != 3 {
		fmt.Printf("specify <config file path> <materials dir path> <storage dir path>\n")
		os.Exit(1)
	}

	confPath := initFlagset.Arg(0)
	materials := initFlagset.Arg(1)
	storagePath := initFlagset.Arg(2)

	cfi, err := os.Stat(confPath)
	if err != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", confPath, err)
		os.Exit(1)
	}

	if !cfi.Mode().IsRegular() {
		fmt.Printf("config file (%s) must be a regular file\n", confPath)
		os.Exit(1)
	}
	if _initFlags.Force {
		// avoid catastrophe caused by a typo
		wellKnownDirs := map[string] bool {
			".": true, "..": true, "/":true, "/home":true, "/tmp":true, "/dummyWellKnownDir":true }
		if wellKnownDirs[storagePath] {
			fmt.Printf("storage dir(%s) typo?\n", storagePath)
			os.Exit(1)
		}
		if err = os.RemoveAll(storagePath); err != nil {
			// NOTE: if storagePath does not exist, RemoveAll returns nil
			fmt.Printf("could not remove %s (%s)\n", storagePath, err)
			os.Exit(1)
		}
		if err = os.Mkdir(storagePath, 0777); err != nil {
			fmt.Printf("could not create %s (%s)\n", storagePath, err)
			os.Exit(1)
		}
	}

	sfi, err := os.Stat(storagePath)
	if err != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", storagePath, err)
		os.Exit(1)
	}

	if !sfi.Mode().IsDir() {
		fmt.Printf("storagePath directory (%s) must be a directory\n", storagePath)
		os.Exit(1)
	}

	dir, err := os.Open(storagePath)
	if err != nil {
		fmt.Printf("failed to open storagePath directory: %s (%s)\n", storagePath, err)
		os.Exit(1)
	}

	fi, err := dir.Readdir(0)
	if err != nil {
		fmt.Printf("failed to read storagePath directory: %s (%s)\n", storagePath, err)
		os.Exit(1)
	}

	if len(fi) != 0 {
		fmt.Printf("directory for earthquake storagePath (%s) must be empty\n", storagePath)
		os.Exit(1)
	}

	config, err := ParseConfigFile(confPath)
	if err != nil {
		fmt.Printf("parsing config file (%s) failed: %s\n", confPath, err)
		os.Exit(1)
	}

	// conf may be JSON/YAML/TOML, but always converted to JSON
	err = config.DumpToJsonFile(path.Join(storagePath, historystorage.StorageConfigPath))
	if err != nil {
		fmt.Printf("placing config file (%s) failed (%s)\n", confPath, err)
		os.Exit(1)
	}

	materialsDir := path.Join(storagePath, storageMaterialsPath)
	err = os.Mkdir(materialsDir, 0777)
	if err != nil {
		fmt.Printf("creating a directory for materials (%s) failed (%s)\n",
			materialsDir, err)
		os.Exit(1)
		// TODO: cleaning conf file
	}

	err = recursiveHardLink(materials, materialsDir)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	storage := historystorage.New(config.GetString("storageType"), storagePath)
	storage.CreateStorage()

	if config.GetString("init") != "" {
		initScriptPath := materialsDir + "/" + config.GetString("init")
		runCmd := createCmd(initScriptPath, "", materialsDir)
		if err = runCmd.Run(); err != nil {
			fmt.Printf("could not run %s (%s)\n", initScriptPath, err)
			os.Exit(1)
		}
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

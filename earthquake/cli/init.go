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
	"flag"
	"fmt"
	"os"
	"path"

	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/osrg/earthquake/earthquake/historystorage"
	"github.com/osrg/earthquake/earthquake/util/cmd"
	"github.com/osrg/earthquake/earthquake/util/config"
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
			if fi.Mode()&os.ModeSymlink != 0 {
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

// FIXME: refactor
func _init(args []string) int {
	if err := initFlagset.Parse(args); err != nil {
		fmt.Printf("%s", err.Error())
		return 1
	}

	if initFlagset.NArg() != 3 {
		fmt.Printf("specify <config file path> <materials dir path> <storage dir path>\n")
		return 1
	}

	confPath := initFlagset.Arg(0)
	materials := initFlagset.Arg(1)
	storagePath := initFlagset.Arg(2)

	cfi, err := os.Stat(confPath)
	if err != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", confPath, err)
		return 1
	}

	if !cfi.Mode().IsRegular() {
		fmt.Printf("config file (%s) must be a regular file\n", confPath)
		return 1
	}
	if _initFlags.Force {
		// avoid catastrophe caused by a typo
		wellKnownDirs := map[string]bool{
			".": true, "..": true, "/": true, "/home": true, "/tmp": true, "/dummyWellKnownDir": true}
		if wellKnownDirs[storagePath] {
			fmt.Printf("storage dir(%s) typo?\n", storagePath)
			return 1
		}
		if err = os.RemoveAll(storagePath); err != nil {
			// NOTE: if storagePath does not exist, RemoveAll returns nil
			fmt.Printf("could not remove %s (%s)\n", storagePath, err)
			return 1
		}
		if err = os.Mkdir(storagePath, 0777); err != nil {
			fmt.Printf("could not create %s (%s)\n", storagePath, err)
			return 1
		}
	}

	sfi, err := os.Stat(storagePath)
	if err != nil {
		fmt.Printf("failed to stat path: %s (%s)\n", storagePath, err)
		return 1
	}

	if !sfi.Mode().IsDir() {
		fmt.Printf("storagePath directory (%s) must be a directory\n", storagePath)
		return 1
	}

	dir, err := os.Open(storagePath)
	if err != nil {
		fmt.Printf("failed to open storagePath directory: %s (%s)\n", storagePath, err)
		return 1
	}

	fi, err := dir.Readdir(0)
	if err != nil {
		fmt.Printf("failed to read storagePath directory: %s (%s)\n", storagePath, err)
		return 1
	}

	if len(fi) != 0 {
		fmt.Printf("directory for earthquake storagePath (%s) must be empty\n", storagePath)
		return 1
	}

	cfg, err := config.NewFromFile(confPath)
	if err != nil {
		fmt.Printf("parsing config file (%s) failed: %s\n", confPath, err)
		return 1
	}

	if !strings.EqualFold(filepath.Ext(confPath), ".toml") {
		fmt.Printf("this version does not support non-TOML configs")
		return 1
	}
	err = copyFile(path.Join(storagePath, historystorage.StorageTOMLConfigPath), confPath, 0644)
	if err != nil {
		fmt.Printf("placing config file (%s) failed (%s)\n", confPath, err)
		return 1
	}

	materialsDir := path.Join(storagePath, storageMaterialsPath)
	err = os.Mkdir(materialsDir, 0777)
	if err != nil {
		fmt.Printf("creating a directory for materials (%s) failed (%s)\n",
			materialsDir, err)
		return 1
		// TODO: cleaning conf file
	}
	cmd.DefaultFactory.SetMaterialsDir(materialsDir)

	err = recursiveHardLink(materials, materialsDir)
	if err != nil {
		fmt.Printf("%s\n", err)
		return 1
	}

	storage, err := historystorage.New(cfg.GetString("storageType"), storagePath)
	if err != nil {
		fmt.Printf("%s\n", err)
		return 1
	}
	storage.CreateStorage()

	if cfg.GetString("init") != "" {
		initScriptPath := path.Join(materialsDir, cfg.GetString("init"))
		runCmd := cmd.DefaultFactory.CreateCmd(initScriptPath)
		if err = runCmd.Run(); err != nil {
			fmt.Printf("could not run %s (%s)\n", initScriptPath, err)
			return 1
		}
	}
	fmt.Printf("ok\n")
	return 0
}

func copyFile(dst, src string, perm os.FileMode) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, content, perm)
	if err != nil {
		return err
	}
	return nil
}

type initCmd struct {
}

func (cmd initCmd) Help() string {
	return "init help (todo)"
}

func (cmd initCmd) Run(args []string) int {
	return _init(args)
}

func (cmd initCmd) Synopsis() string {
	return "Initialize storage directory"
}

func initCommandFactory() (cli.Command, error) {
	return initCmd{}, nil
}

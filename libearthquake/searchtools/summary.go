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

package searchtools

import (
	"fmt"
	// "io/ioutil"
	"os"
	// "sort"

	// . "../equtils"
	"../historystorage"

	"github.com/mitchellh/cli"
)

func doSummary(historyStoragePath string) {
	// FIXME: storage type
	storage := historystorage.New("naive", historyStoragePath)

	storage.Init()
	nrStored := storage.NrStoredHistories()

	for i := 0; i < nrStored; i++ {
		succeed, err := storage.IsSucceed(i)
		if err != nil {
			fmt.Printf("failed to open history %08x, %s\n", i, err)
			continue
		}

		if !succeed {
			fmt.Printf("%08x caused failure\n", i)
		}
	}
}

func summary(args []string) {
	if len(args) != 1 {
		fmt.Printf("need history storage path\n")
		os.Exit(1)
	}

	doSummary(args[0])
}

type summaryCmd struct {
}

func (cmd summaryCmd) Help() string {
	return "summary help (todo)"
}

func (cmd summaryCmd) Run(args []string) int {
	summary(args)
	return 0
}

func (cmd summaryCmd) Synopsis() string {
	return "summary subcommand"
}

func SummaryCommandFactory() (cli.Command, error) {
	return summaryCmd{}, nil
}

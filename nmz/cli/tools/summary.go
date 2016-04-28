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

package tools

import (
	"flag"
	"fmt"
	"time"

	"github.com/mitchellh/cli"
	"github.com/osrg/namazu/nmz/historystorage"
)

type summaryFlags struct {
	ListUpOverAverage bool
}

var (
	summaryFlagset = flag.NewFlagSet("summary", flag.ExitOnError)
	_summaryFlags  = summaryFlags{}
)

func init() {
	summaryFlagset.BoolVar(&_summaryFlags.ListUpOverAverage, "list-up-over-average", false, "list up IDs of runs whose time is longer than average")
}

func doSummary(historyStoragePath string) {
	storage := historystorage.LoadStorage(historyStoragePath)

	storage.Init()
	nrStored := storage.NrStoredHistories()

	for i := 0; i < nrStored; i++ {
		succeed, err := storage.IsSuccessful(i)
		if err != nil {
			fmt.Printf("failed to open history %08x, %s\n", i, err)
			continue
		}

		if !succeed {
			fmt.Printf("%08x caused failure\n", i)
		}
	}
}

func listUpOverAverage(historyStoragePath string) {
	storage := historystorage.LoadStorage(historyStoragePath)

	storage.Init()
	nrStored := storage.NrStoredHistories()

	totalTime := time.Duration(0)

	for i := 0; i < nrStored; i++ {
		time, err := storage.GetRequiredTime(i)
		if err != nil {
			fmt.Printf("failed to open history %08x, %s\n", i, err)
			continue // just skip?
		}

		totalTime += time
	}

	averageTime := time.Duration(int64(totalTime) / int64(nrStored))

	for i := 0; i < nrStored; i++ {
		time, err := storage.GetRequiredTime(i)
		if err != nil {
			fmt.Printf("failed to open history %08x, %s\n", i, err)
			continue // just skip?
		}

		if averageTime < time {
			fmt.Printf("%08x\n", i)
		}
	}
}

type summaryCmd struct {
}

func SummaryCommandFactory() (cli.Command, error) {
	return summaryCmd{}, nil
}

func (cmd summaryCmd) Synopsis() string {
	return "summary subcommand"
}

func (cmd summaryCmd) Help() string {
	return "Please run `nmz --help tools` instead"
}

func (cmd summaryCmd) Run(args []string) int {
	summaryFlagset.Parse(args)

	if summaryFlagset.NArg() != 1 {
		fmt.Printf("need history storage path\n")
		return 1
	}

	if _summaryFlags.ListUpOverAverage {
		listUpOverAverage(args[len(args)-1])
	} else {
		doSummary(args[len(args)-1])
	}
	return 0
}

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
	"flag"

	. "../equtils"
	"../historystorage"

	"github.com/mitchellh/cli"
)

type visualizeFlags struct {
	Mode string
}

var (
	visualizeFlagset = flag.NewFlagSet("visualize", flag.ExitOnError)
	_visualizeFlags  = visualizeFlags{}
)

func init() {
	visualizeFlagset.StringVar(&_visualizeFlags.Mode, "mode", "", "mode of visualization")
}

func seenBefore(traces []*SingleTrace, newTrace *SingleTrace) bool {
	for _, trace := range traces {
		if len(trace.EventSequence) != len(newTrace.EventSequence) {
			continue
		}

		seen := true

		for i, ev := range trace.EventSequence {
			if !areEventsEqual(&ev, &newTrace.EventSequence[i]) {
				seen = false
				break
			}
		}

		if !seen {
			continue
		}

		return true
	}

	return false
}

func gnuplot(historyStoragePath string) {
	storage := historystorage.LoadStorage(historyStoragePath)

	storage.Init()
	nrStored := storage.NrStoredHistories()
	nrUniques := 0
	uniqueTraces := make([]*SingleTrace, 0)

	for i := 0; i < nrStored; i++ {
		trace, err := storage.GetStoredHistory(i)
		if err != nil {
			fmt.Printf("failed to open history %08x, %s\n", i, err)
			os.Exit(1)
		}

		if !seenBefore(uniqueTraces, trace) {
			nrUniques++
			uniqueTraces = append(uniqueTraces, trace)
		}

		fmt.Printf("%d %d\n", i + 1, nrUniques)
	}
}

func visualize(args []string) {
	visualizeFlagset.Parse(args)

	switch _visualizeFlags.Mode {
	case "gnuplot":
		if visualizeFlagset.NArg() != 1 {
			fmt.Printf("need a path of history storage")
		}
		gnuplot(args[len(args)-1])
	default:
		fmt.Printf("unknown mode of visualize: %s\n", _visualizeFlags.Mode)
	}
}

type visualizeCmd struct {
}

func (cmd visualizeCmd) Help() string {
	return "visualize help (todo)"
}

func (cmd visualizeCmd) Run(args []string) int {
	visualize(args)
	return 0
}

func (cmd visualizeCmd) Synopsis() string {
	return "visualize subcommand"
}

func VisualizeCommandFactory() (cli.Command, error) {
	return visualizeCmd{}, nil
}

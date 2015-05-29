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
	"encoding/gob"
	"fmt"
	"os"
	"flag"

	. "../equtils"

	"github.com/mitchellh/cli"
)

type permutateFlags struct {
	TracePath string
}

var (
	permutateFlagset = flag.NewFlagSet("permutate", flag.ExitOnError)
	_permutateFlags  = permutateFlags{}
)

func init() {
	permutateFlagset.StringVar(&_permutateFlags.TracePath, "trace-path", "", "path of trace data file")
}

func appearedInArray(name string, array []string) bool {
	for _, n := range array {
		if n == name {
			return true
		}
	}

	return false
}

func listUpProcs(trace *SingleTrace) []string {
	nodes := make([]string, 0)

	for _, ev := range trace.EventSequence {
		if appearedInArray(ev.ProcId, nodes) {
			continue
		}

		nodes = append(nodes, ev.ProcId)
	}

	return nodes
}

func listEventsOfProc(proc string, trace *SingleTrace) []Event {
	events := make([]Event, 0)

	for _, ev := range trace.EventSequence {
		if ev.ProcId != proc {
			continue
		}

		events = append(events, ev)
	}

	return events
}

func copyMap(proc2ev map[string][]Event) map[string][]Event {
	ret := make(map[string][]Event)

	for k, v := range proc2ev {
		ret[k] = make([]Event, len(v))
		copy(ret[k], v)
	}

	return ret
}

func isEmptyMap(proc2ev map[string][]Event) bool {
	for _, v := range proc2ev {
		if len(v) != 0 {
			return false
		}
	}

	return true
}

func doPermutate(proc2ev map[string][]Event, constructing []Event, result chan []Event, endCh chan interface{}, level int) {
	if isEmptyMap(proc2ev) {
		result <- constructing
		return
	}

	for proc, ev := range proc2ev {
		if len(ev) == 0 {
			continue
		}

		growed := make([]Event, len(constructing))
		copy(growed, constructing)
		growed = append(growed, ev[0])

		newProc2Ev := copyMap(proc2ev)
		newProc2Ev[proc] = newProc2Ev[proc][1:]

		doPermutate(newProc2Ev, growed, result, endCh, level+1)
	}

	if level == 0 {
		endCh <- true
	}
}

func permutate(args []string) {
	permutateFlagset.Parse(args)

	if _permutateFlags.TracePath == "" {
		fmt.Printf("specify path of trace data file\n")
		os.Exit(1)
	}

	tracepath := _permutateFlags.TracePath
	file, err := os.Open(tracepath)
	if err != nil {
		fmt.Printf("failed to open trace data file(%s): %s\n", tracepath, err)
		os.Exit(1)
	}

	dec := gob.NewDecoder(file)
	var trace SingleTrace
	derr := dec.Decode(&trace)
	if derr != nil {
		fmt.Printf("failed to decode trace file(%s): %s\n", tracepath, err)
		os.Exit(1)
	}

	procs := listUpProcs(&trace)

	eventsPerProc := make(map[string][]Event)

	for _, proc := range procs {
		eventsPerProc[proc] = listEventsOfProc(proc, &trace)
	}

	ch := make(chan []Event)
	endCh := make(chan interface{})
	constructing := make([]Event, 0)

	go doPermutate(eventsPerProc, constructing, ch, endCh, 0)

	result := make([][]Event, 0)

	run := true
	for run {
		select {
		case newPattern := <-ch:
			result = append(result, newPattern)
		case <-endCh:
			run = false
		}
	}

	for _, r := range result {
		for _, e := range r {
			fmt.Printf("%s: %s(%s), ", e.ProcId, e.EventType, e.EventParam)
		}
		fmt.Printf("\n")
	}
}

type permutateCmd struct {
}

func (cmd permutateCmd) Help() string {
	return "permutate help (todo)"
}

func (cmd permutateCmd) Run(args []string) int {
	permutate(args)
	return 0
}

func (cmd permutateCmd) Synopsis() string {
	return "permutate subcommand"
}

func PermutateCommandFactory() (cli.Command, error) {
	return permutateCmd{}, nil
}

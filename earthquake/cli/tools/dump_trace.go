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
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/mitchellh/cli"
	. "github.com/osrg/earthquake/earthquake/signal"
	. "github.com/osrg/earthquake/earthquake/util/trace"
)

type dumpTraceFlags struct {
	TracePath string
}

var (
	dumpTraceFlagset = flag.NewFlagSet("dump-trace", flag.ExitOnError)
	_dumpTraceFlags  = dumpTraceFlags{}
)

func init() {
	dumpTraceFlagset.StringVar(&_dumpTraceFlags.TracePath, "trace-path", "", "path of trace data file")
}

func dumpMap(m map[string]interface{}, desc string) {
	fmt.Printf(" - %s:\n", desc)
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf(" -- %s: %#v\n", k, m[k])
	}
}

func dumpActionDetails(action Action) {
	intf := action.JSONMap()["option"]
	m, ok := intf.(map[string]interface{})
	if ok && len(m) > 0 {
		dumpMap(m, "Action option")
	}
}

func dumpEventDetails(event Event) {
	intf := event.JSONMap()["option"]
	m, ok := intf.(map[string]interface{})
	if ok && len(m) > 0 {
		dumpMap(m, "Event option")
	}
}

func dumpActionWithoutEvent(i int, action Action) {
	fmt.Printf("%d: %s[%s@%s]\n",
		i,
		action.JSONMap()["class"], action.EntityID(), action.TriggeredTime().Local().Format(time.StampMicro))
	dumpActionDetails(action)
}

func dumpAction(i int, action Action) {
	event := action.Event()
	if event == nil {
		dumpActionWithoutEvent(i, action)
		return
	}
	fmt.Printf("%d: %s[%s@%s] for %s[%s@%s]\n",
		i,
		action.JSONMap()["class"], action.EntityID(), action.TriggeredTime().Local().Format(time.StampMicro),
		event.JSONMap()["class"], event.EntityID(), event.ArrivedTime().Local().Format(time.StampMicro))
	dumpActionDetails(action)
	dumpEventDetails(event)
}

func doDumpTrace(trace *SingleTrace) {
	for i, act := range trace.ActionSequence {
		dumpAction(i, act)
	}
}

type dumpTraceCmd struct {
}

func DumpTraceCommandFactory() (cli.Command, error) {
	return dumpTraceCmd{}, nil
}

func (cmd dumpTraceCmd) Synopsis() string {
	return "dumpTrace subcommand"
}

func (cmd dumpTraceCmd) Help() string {
	return "Please run `earthquake --help tools` instead"
}

func (cmd dumpTraceCmd) Run(args []string) int {
	dumpTraceFlagset.Parse(args)

	if _dumpTraceFlags.TracePath == "" {
		fmt.Printf("specify path of trace data file\n")
		return 1
	}

	file, err := os.Open(_dumpTraceFlags.TracePath)
	if err != nil {
		fmt.Printf("failed to open trace data file(%s): %s\n", _dumpTraceFlags.TracePath, err)
		return 1
	}
	var trace SingleTrace
	if err = gob.NewDecoder(file).Decode(&trace); err != nil {
		fmt.Printf("failed to decode trace file(%s): %s\n", _dumpTraceFlags.TracePath, err)
		return 1
	}

	doDumpTrace(&trace)
	return 0
}

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
	. "../equtils"
	"encoding/gob"
	"flag"
	"fmt"
	"encoding/json"
	"os"
	"time"
	"github.com/mitchellh/cli"
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

func dumpJSONAction(i int, act *Action) {
	if ( act.ActionType != "_JSON") {
		panic(fmt.Errorf("bad action %s", act))
	}
	ev := act.Evt
	fmt.Printf("%d %s for @ %s: %s, %s\n",
		i,
		act.ActionParam["class"],
		ev.ArrivedTime.Local().Format(time.UnixDate), ev.ProcId, ev.EventParam["class"])
	actJson, err := json.MarshalIndent(act.ActionParam, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", actJson)
	evJson, err := json.MarshalIndent(ev.EventParam, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", evJson)
}

func dumpAction(i int, act *Action) {
	if (act.ActionType == "_JSON") {
		dumpJSONAction(i, act)
		return
	}
	ev := act.Evt
	fmt.Printf("%d %s(%s) for @ %s: %s, %s(%s)\n",
		i,
		act.ActionType, act.ActionParam,
		ev.ArrivedTime.Local().Format(time.UnixDate), ev.ProcId, ev.EventType, ev.EventParam)

	if ev.JavaSpecific != nil {
		js := ev.JavaSpecific

		fmt.Printf("\tThread: %s\n", js.ThreadName)

		fmt.Printf("\tparams:\n")
		for _, param := range js.Params {
			fmt.Printf("\t\t%s: %s\n", param.Name, param.Value)
		}

		fmt.Printf("\tstack trace:\n")
		for _, stackTraceElement := range js.StackTraceElements {
			fmt.Printf("\t\t%s %s %s %d\n",
				stackTraceElement.FileName,
				stackTraceElement.ClassName,
				stackTraceElement.MethodName,
				stackTraceElement.LineNumber)
		}
	}
}

func doDumpTrace(trace *SingleTrace) {
	for i, act := range trace.ActionSequence {
		dumpAction(i, &act)
	}
}

func dumpTrace(args []string) {
	dumpTraceFlagset.Parse(args)

	if _dumpTraceFlags.TracePath == "" {
		fmt.Printf("specify path of trace data file\n")
		os.Exit(1)
	}

	file, err := os.Open(_dumpTraceFlags.TracePath)
	if err != nil {
		fmt.Printf("failed to open trace data file(%s): %s\n", _dumpTraceFlags.TracePath, err)
		os.Exit(1)
	}
	gob.Register(map[string]interface{}{})
	gob.Register([]map[string]interface{}{})
	dec := gob.NewDecoder(file)
	var trace SingleTrace
	derr := dec.Decode(&trace)
	if derr != nil {
		fmt.Printf("failed to decode trace file(%s): %s\n", _dumpTraceFlags.TracePath, err)
		os.Exit(1)
	}

	doDumpTrace(&trace)
}

type dumpTraceCmd struct {
}

func (cmd dumpTraceCmd) Help() string {
	return "dumpTrace help (todo)"
}

func (cmd dumpTraceCmd) Run(args []string) int {
	dumpTrace(args)
	return 0
}

func (cmd dumpTraceCmd) Synopsis() string {
	return "dumpTrace subcommand"
}

func DumpTraceCommandFactory() (cli.Command, error) {
	return dumpTraceCmd{}, nil
}

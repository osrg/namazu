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
	"encoding/gob"
	"flag"
	"fmt"
	"os"
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

	dec := gob.NewDecoder(file)
	var trace SingleTrace
	derr := dec.Decode(&trace)
	if derr != nil {
		fmt.Printf("failed to decode trace file(%s): %s\n", _dumpTraceFlags.TracePath, err)
		os.Exit(1)
	}

	for i, ev := range trace.EventSequence {
		fmt.Printf("%d: %s, %s(%s)\n", i, ev.ProcId, ev.EventType, ev.EventParam)
	}
}

func runSearchTools(name string, args []string) {
	switch name {
	case "dump-trace":
		dumpTrace(args)
	default:
		fmt.Printf("unknown subcommand: %s\n", name)
		os.Exit(1)
	}
}

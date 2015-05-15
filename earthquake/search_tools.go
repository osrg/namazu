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
	"flag"
	"fmt"
	"os"
)

type dumpTraceFlags struct {
	TracePath string
}

type calcDuplicationFlags struct {
	TraceDir string
}

type visualizeFlags struct {
	Mode     string
	TraceDir string
}

type permutateFlags struct {
	TracePath string
}

var (
	dumpTraceFlagset = flag.NewFlagSet("dump-trace", flag.ExitOnError)
	_dumpTraceFlags  = dumpTraceFlags{}

	calcDuplicationFlagset = flag.NewFlagSet("calc-duplication", flag.ExitOnError)
	_calcDuplicationFlags  = calcDuplicationFlags{}

	visualizeFlagset = flag.NewFlagSet("visualize", flag.ExitOnError)
	_visualizeFlags  = visualizeFlags{}

	permutateFlagset = flag.NewFlagSet("permutate", flag.ExitOnError)
	_permutateFlags  = permutateFlags{}
)

func init() {
	dumpTraceFlagset.StringVar(&_dumpTraceFlags.TracePath, "trace-path", "", "path of trace data file")

	calcDuplicationFlagset.StringVar(&_calcDuplicationFlags.TraceDir, "trace-dir", "", "path of trace data directory")

	visualizeFlagset.StringVar(&_visualizeFlags.TraceDir, "trace-dir", "", "path of trace data directory")
	visualizeFlagset.StringVar(&_visualizeFlags.Mode, "mode", "", "mode of visualization")

	permutateFlagset.StringVar(&_permutateFlags.TracePath, "trace-path", "", "path of trace data file")
}

func runSearchTools(name string, args []string) {
	switch name {
	case "dump-trace":
		dumpTrace(args)
	case "calc-duplication":
		calcDuplication(args)
	case "visualize":
		visualize(args)
	case "permutate":
		permutate(args)
	default:
		fmt.Printf("unknown subcommand: %s\n", name)
		os.Exit(1)
	}
}

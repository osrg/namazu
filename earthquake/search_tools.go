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
	"io/ioutil"
	"os"
)

type dumpTraceFlags struct {
	TracePath string
}

type calcDuplicationFlags struct {
	TraceDir string
}

var (
	dumpTraceFlagset = flag.NewFlagSet("dump-trace", flag.ExitOnError)
	_dumpTraceFlags  = dumpTraceFlags{}

	calcDuplicationFlagset = flag.NewFlagSet("calc-duplication", flag.ExitOnError)
	_calcDuplicationFlags  = calcDuplicationFlags{}
)

func init() {
	dumpTraceFlagset.StringVar(&_dumpTraceFlags.TracePath, "trace-path", "", "path of trace data file")

	calcDuplicationFlagset.StringVar(&_calcDuplicationFlags.TraceDir, "trace-dir", "", "path of trace data directory")
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

func calcDuplication(args []string) {
	calcDuplicationFlagset.Parse(args)

	traceDir := _calcDuplicationFlags.TraceDir
	if traceDir == "" {
		fmt.Printf("specify directory path of trace data\n")
		os.Exit(1)
	}

	des, err := ioutil.ReadDir(traceDir)
	if err != nil {
		fmt.Printf("failed to read directory(%s): %s\n", traceDir, err)
		os.Exit(1)
	}

	if len(des) == 0 {
		fmt.Printf("directory %s is empty\n", traceDir)
		os.Exit(1)
	}

	infoFile, oerr := os.Open(traceDir + "/" + SearchModeInfoPath)
	if oerr != nil {
		fmt.Printf("failed to read info file of search mode: %s\n", oerr)
		os.Exit(1)
	}

	infoDec := gob.NewDecoder(infoFile)
	var info SearchModeInfo
	derr := infoDec.Decode(&info)
	if derr != nil {
		fmt.Printf("failed to decode info file: %s\n", derr)
		os.Exit(1)
	}

	fmt.Printf("a number of collected traces: %d\n", info.NrCollectedTraces)
}

func runSearchTools(name string, args []string) {
	switch name {
	case "dump-trace":
		dumpTrace(args)
	case "calc-duplication":
		calcDuplication(args)
	default:
		fmt.Printf("unknown subcommand: %s\n", name)
		os.Exit(1)
	}
}

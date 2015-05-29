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
	"io/ioutil"
	"os"
	"flag"

	. "../equtils"

	"github.com/mitchellh/cli"
)

type calcDuplicationFlags struct {
	TraceDir string
}

var (
	calcDuplicationFlagset = flag.NewFlagSet("calc-duplication", flag.ExitOnError)
	_calcDuplicationFlags  = calcDuplicationFlags{}
)

func init() {
	calcDuplicationFlagset.StringVar(&_calcDuplicationFlags.TraceDir, "trace-dir", "", "path of trace data directory")
}

func areEventsEqual(a, b *Event) bool {
	if a.ProcId != b.ProcId {
		return false
	}

	if a.EventType != b.EventType {
		return false
	}

	if a.EventParam != b.EventParam {
		return false
	}

	return true
}

func areTracesEqual(a, b *SingleTrace) bool {
	if len(a.EventSequence) != len(b.EventSequence) {
		return false
	}

	for i := 0; i < len(a.EventSequence); i++ {
		if !areEventsEqual(&a.EventSequence[i], &b.EventSequence[i]) {
			return false
		}
	}

	return true
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

	traces := make([]*SingleTrace, info.NrCollectedTraces)
	for i := 0; i < info.NrCollectedTraces; i++ {
		path := fmt.Sprintf("%s/%08x", traceDir, i)
		traceFile, toerr := os.Open(path)
		if toerr != nil {
			fmt.Printf("failed to open trace file(%s): %s\n", traceFile, toerr)
			os.Exit(1)
		}

		tDec := gob.NewDecoder(traceFile)
		var newTrace SingleTrace
		derr = tDec.Decode(&newTrace)
		if derr != nil {
			fmt.Printf("decoding file(%s) failed: %s\n", path, derr)
			os.Exit(1)
		}

		traces[i] = &newTrace
	}

	equalRelation := make([][]int, info.NrCollectedTraces)
	for i := 0; i < info.NrCollectedTraces; i++ {
		equalRelation[i] = make([]int, 0)
	}

	for i := 0; i < info.NrCollectedTraces; i++ {
		for j := 0; j < i; j++ {
			if areTracesEqual(traces[i], traces[j]) {
				equalRelation[i] = append(equalRelation[i], j)
			}
		}
	}

	nrUniqueTraces := 0
	for i := 0; i < info.NrCollectedTraces; i++ {
		// fmt.Printf("%d: %d\n",i , len(equalRelation[i]))
		if len(equalRelation[i]) == 0 {
			nrUniqueTraces++
		}
	}

	fmt.Printf("a number of unique traces: %d\n", nrUniqueTraces)
	fmt.Printf("a number of entire traces: %d\n", info.NrCollectedTraces)
}

type calcDupCmd struct {
}

func (cmd calcDupCmd) Help() string {
	return "calcDup help (todo)"
}

func (cmd calcDupCmd) Run(args []string) int {
	calcDuplication(args)
	return 0
}

func (cmd calcDupCmd) Synopsis() string {
	return "calcDup subcommand"
}

func CalcDupCommandFactory() (cli.Command, error) {
	return calcDupCmd{}, nil
}

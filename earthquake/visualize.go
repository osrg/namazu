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
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

type orderedEvents []*Event

func (e orderedEvents) Len() int {
	return len(e)
}

func (e orderedEvents) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e orderedEvents) Less(i, j int) bool {
	if e[i].ProcId != e[j].ProcId {
		return e[i].ProcId < e[j].ProcId
	}

	if e[i].EventType != e[j].EventType {
		return e[i].EventType < e[j].EventType
	}

	return e[i].EventParam < e[j].EventParam
}

type state struct {
	events orderedEvents
	trans  []*transition
}

type transition struct {
	destination *state
	count       int
}

var foundStates []*state

func singleTraceToState(seq []Event) *state {
	var s state

	for _, e := range seq {
		s.events = append(s.events, &e)
	}

	sort.Sort(s.events)
	s.trans = make([]*transition, 0)

	return &s
}

func areStatesEqual(a, b *state) bool {
	if len(a.events) != len(b.events) {
		return false
	}

	for i, _ := range a.events {
		if !areEventsEqual(a.events[i], b.events[i]) {
			return false
		}
	}

	return true
}

func isStateFound(s *state) bool {
	for _, _s := range foundStates {
		if areStatesEqual(_s, s) {
			return true
		}
	}

	return false
}

func incTransition(src, dst *state) {
	for _, t := range src.trans {
		if t.destination == dst {
			t.count++
			return
		}
	}

	newTrans := transition{
		dst, 0,
	}

	src.trans = append(src.trans, &newTrans)
}

func visualize(args []string) {
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

		var prevState *state
		prevState = nil
		for i, _ := range newTrace.EventSequence {
			if i == 0 {
				// skip empty trace
				continue
			}

			t := newTrace.EventSequence[0:i]
			s := singleTraceToState(t)
			if isStateFound(s) {
				continue
			}

			foundStates = append(foundStates, s)

			if prevState == nil {
				prevState = s
				continue
			}

			incTransition(prevState, s)
			prevState = s
		}
	}

	for _, s := range foundStates {
		fmt.Printf("%v\n", s)
	}

}

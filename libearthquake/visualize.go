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
	"encoding/json"
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

func areOrderedEventsEqual(a, b orderedEvents) bool {
	if len(a) != len(b) {
		return false
	}

	for i, ae := range a {
		if !areEventsEqual(ae, b[i]) {
			return false
		}
	}

	return true
}

func incTransition(src, dst *state) {
	for _, t := range src.trans {
		if areOrderedEventsEqual(t.destination.events, dst.events) {
			t.count++
			return
		}
	}

	newTrans := transition{
		dst, 1,
	}

	src.trans = append(src.trans, &newTrans)
}

func indexLookup(s *state) int {
	for i, _s := range foundStates {
		if _s == s {
			return i
		}
	}

	return -1
}

func d3(traceDir string) {
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

	initEvents := make([]*Event, 0)
	initTrans := make([]*transition, 0)
	initState := state{
		initEvents,
		initTrans,
	}
	foundStates = append(foundStates, &initState)

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

		prevState := &initState
		for i, _ := range newTrace.EventSequence {
			if i == 0 {
				// skip empty trace
				continue
			}

			t := newTrace.EventSequence[0:i]
			s := singleTraceToState(t)

			incTransition(prevState, s)

			if isStateFound(s) {
				prevState = s
				continue
			}

			foundStates = append(foundStates, s)

			prevState = s
		}
	}

	fmt.Printf("{\n")

	// output nodes first
	fmt.Printf("\"nodes\":[\n")
	for i, s := range foundStates {
		fmt.Printf("{\"name\": \"%p\", \"group\": %d}", s, i)
		if i != len(foundStates)-1 {
			fmt.Printf(",\n")
		}
	}
	fmt.Printf("],\n")

	// output links
	fmt.Printf("\"links\":\n")

	type linkUnit struct {
		Source int `json:"source"`
		Target int `json:"target"`
		Value  int `json:"value"`
	}
	links := make([]*linkUnit, 0)

	for i, src := range foundStates {
		if len(src.trans) == 0 {
			continue
		}

		for _, t := range src.trans {
			newLink := linkUnit{
				i, indexLookup(t.destination), t.count,
			}

			links = append(links, &newLink)
		}
	}

	encoded, jerr := json.Marshal(links)
	if jerr != nil {
		fmt.Printf("encoding error: %s\n", jerr)
		os.Exit(1)
	}
	fmt.Printf(string(encoded))

	fmt.Printf("}\n")

}

func isSeenTrace(tr *SingleTrace, traces []*SingleTrace) bool {
	for _, seen := range traces {
		if len(tr.EventSequence) != len(seen.EventSequence) {
			return false
		}

		for i, ev := range tr.EventSequence {
			if !areEventsEqual(&ev, &seen.EventSequence[i]) {
				return false
			}
		}

		return true
	}

	return false
}

func gnuplot(traceDir string) {
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

	pastTraces := make([]*SingleTrace, 0)
	nrUniqueEvents := 0

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

		if !isSeenTrace(&newTrace, pastTraces) {
			nrUniqueEvents++
			pastTraces = append(pastTraces, &newTrace)
		}

		fmt.Printf("%d %d\n", i, nrUniqueEvents)

		traceFile.Close()
	}
}

func visualize(args []string) {
	visualizeFlagset.Parse(args)

	switch _visualizeFlags.Mode {
	case "d3js":
		d3(_visualizeFlags.TraceDir)
	case "gnuplot":
		gnuplot(_visualizeFlags.TraceDir)
	default:
		fmt.Printf("unknown mode of visualize: %s\n", _visualizeFlags.Mode)
	}
}

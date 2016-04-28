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
	"flag"
	"fmt"
	"sort"

	"github.com/mitchellh/cli"
	"github.com/osrg/namazu/nmz/historystorage"
	. "github.com/osrg/namazu/nmz/signal"
	. "github.com/osrg/namazu/nmz/util/trace"
)

type visualizeFlags struct {
	Mode string

	POReduction bool
}

var (
	visualizeFlagset = flag.NewFlagSet("visualize", flag.ExitOnError)
	_visualizeFlags  = visualizeFlags{}
)

func init() {
	visualizeFlagset.StringVar(&_visualizeFlags.Mode, "mode", "", "mode of visualization")
	visualizeFlagset.BoolVar(&_visualizeFlags.POReduction, "po-reduction", true, "count with partial order reduction")
}

type uniqueTraceUnit struct {
	trace *SingleTrace

	tracesPerTransitionEntity map[string][]Event
}

func seenBefore(traces []*uniqueTraceUnit, newUnit *uniqueTraceUnit) bool {
	for _, _trace := range traces {
		trace := _trace.trace
		newTrace := newUnit.trace
		if trace.Equals(newTrace) {
			return true
		}
	}
	return false
}

func createTracesPerEntity(_trace *uniqueTraceUnit) {
	trace := _trace.trace

	perEntity := make(map[string][]Event)

	for _, act := range trace.ActionSequence {
		evt := act.Event()
		if evt != nil {
			entityID := evt.EntityID()
			if _, ok := perEntity[entityID]; !ok {
				perEntity[entityID] = make([]Event, 0)
			}
			perEntity[entityID] = append(perEntity[entityID], evt)
		}
	}

	_trace.tracesPerTransitionEntity = perEntity
}

func tracesEqualInPO(_a, _b *uniqueTraceUnit) bool {
	a := _a.tracesPerTransitionEntity
	b := _b.tracesPerTransitionEntity

	entitiesOfA := make([]string, 0)
	for entity, _ := range a {
		entitiesOfA = append(entitiesOfA, entity)
	}

	entitiesOfB := make([]string, 0)
	for entity, _ := range b {
		entitiesOfB = append(entitiesOfB, entity)
	}

	if len(entitiesOfA) != len(entitiesOfB) {
		return false
	}

	sort.Strings(entitiesOfA)
	sort.Strings(entitiesOfB)

	for i, entity := range entitiesOfA {
		if entitiesOfB[i] != entity {
			return false
		}
	}

	for _, entityOfA := range entitiesOfA {
		aEvents := a[entityOfA]
		bEvents := b[entityOfA]

		if len(aEvents) != len(bEvents) {
			return false
		}

		for i, evt := range aEvents {
			if !evt.Equals(bEvents[i]) {
				return false
			}
		}
	}

	return true
}

func seenBeforePOR(traces []*uniqueTraceUnit, newUnit *uniqueTraceUnit) bool {
	createTracesPerEntity(newUnit)

	for _, trace := range traces {
		if tracesEqualInPO(trace, newUnit) {
			return true
		}
	}

	return false
}

func gnuplot(historyStoragePath string, poReduction bool) error {
	storage := historystorage.LoadStorage(historyStoragePath)

	storage.Init()
	nrStored := storage.NrStoredHistories()
	nrUniques := 0
	uniqueTraces := make([]*uniqueTraceUnit, 0)

	for i := 0; i < nrStored; i++ {
		trace, err := storage.GetStoredHistory(i)
		if err != nil {
			return fmt.Errorf("failed to open history %08x, %s\n", i, err)
		}

		newUnit := &uniqueTraceUnit{
			trace: trace,
		}

		seen := false

		if poReduction {
			seen = seenBeforePOR(uniqueTraces, newUnit)
		} else {
			seen = seenBefore(uniqueTraces, newUnit)
		}

		if !seen {
			nrUniques++
			uniqueTraces = append(uniqueTraces, newUnit)
		}

		fmt.Printf("%d %d\n", i+1, nrUniques)
	}
	return nil
}

type visualizeCmd struct {
}

func VisualizeCommandFactory() (cli.Command, error) {
	return visualizeCmd{}, nil
}

func (cmd visualizeCmd) Synopsis() string {
	return "visualize subcommand"
}

func (cmd visualizeCmd) Help() string {
	return "Please run `nmz --help tools` instead"
}

func (cmd visualizeCmd) Run(args []string) int {
	visualizeFlagset.Parse(args)

	switch _visualizeFlags.Mode {
	case "gnuplot":
		if visualizeFlagset.NArg() != 1 {
			fmt.Printf("need a path of history storage")
		}
		if err := gnuplot(args[len(args)-1], _visualizeFlags.POReduction); err != nil {
			fmt.Printf("%s", err)
			return 1
		}
	default:
		fmt.Printf("unknown mode of visualize: %s\n", _visualizeFlags.Mode)
		return 1
	}
	return 0
}

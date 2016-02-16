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

package random

import (
	"fmt"
	"github.com/AkihiroSuda/go-linuxsched"
	log "github.com/cihub/seelog"
	"github.com/leesper/go_rng"
	"github.com/osrg/earthquake/earthquake/signal"
	"runtime"
	"time"
)

var (
	drng = rng.NewDirichletGenerator(time.Now().UnixNano())
)

func (r *Random) makeActionForProcSetEvent(event *signal.ProcSetEvent) (signal.Action, error) {
	xprocs, ok := event.Option()["procs"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no procs? this should be an implementation error. event=%#v", event)
	}

	// due to JSON nature, we need to convert []interface{} to []string here
	procs := []string{}
	for _, xproc := range xprocs {
		proc, ok := xproc.(string)
		if !ok {
			return nil, fmt.Errorf("non-string %#v", xproc)
		}
		procs = append(procs, proc)
	}

	attrs := r.dirichletSchedDeadline(procs, time.Millisecond, 1.0)
	for pidStr, attr := range attrs {
		log.Infof("For PID=%s, setting Attr=%v", pidStr, attr)
	}
	return signal.NewProcSetSchedAction(event, attrs)
}

// due to JSON nature, we use string for PID representation
func (r *Random) dirichletSchedDeadline(procs []string, base time.Duration, eff float64) map[string]linuxsched.SchedAttr {
	attrs := make(map[string]linuxsched.SchedAttr, len(procs))
	ratios := drng.FlatDirichlet(len(procs))
	for i, pidStr := range procs {
		numCPU := runtime.NumCPU()
		runtime := time.Duration(int(float64(base) * ratios[i] * eff * float64(numCPU)))
		deadline := base
		period := deadline
		attrs[pidStr] = linuxsched.SchedAttr{
			Policy:   linuxsched.Deadline,
			Flags:    linuxsched.ResetOnFork,
			Runtime:  runtime,
			Deadline: deadline,
			Period:   period,
		}
	}
	return attrs
}

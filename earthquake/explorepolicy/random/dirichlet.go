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
	"math/rand"
	"runtime"
	"time"
)

var (
	drng = rng.NewDirichletGenerator(time.Now().UnixNano())
)

type dirichlet struct {
	r *Random
}

// implements procPolicyIntf
func (d *dirichlet) Action(event *signal.ProcSetEvent) (signal.Action, error) {
	procs, err := d.parseProcSetEvent(event)
	if err != nil {
		return nil, err
	}
	attrs := d.dirichletSchedDeadline(procs, time.Millisecond, 1.0)
	for pidStr, attr := range attrs {
		log.Debugf("For PID=%s, setting Attr=%v", pidStr, attr)
	}
	return signal.NewProcSetSchedAction(event, attrs)
}

// parses *ProcSetEvent and returns array of PIDs.
//
// due to JSON nature, we use string for PID representation.
func (d *dirichlet) parseProcSetEvent(event *signal.ProcSetEvent) ([]string, error) {
	option := event.Option()
	procs, ok := option["procs"].([]string)
	if !ok {
		// FIXME: this may not work with REST endpoint.
		// we need to convert []interface{} to []string here..?
		return nil, fmt.Errorf("no procs? this should be an implementation error. event=%#v", event)
	}
	return procs, nil
}

// Returns map of linuxsched.SchedAttr{Policy: linuxsched.Deadline} for procs.
// The runtime value is base * r * eff * numCPU, where r is dirichlet-distributed random numbers.
// (we should improve this strategy.)
//
// due to JSON nature, we use string for PID representation.
func (d *dirichlet) dirichletSchedDeadline(procs []string, base time.Duration, eff float64) map[string]linuxsched.SchedAttr {
	attrs := make(map[string]linuxsched.SchedAttr, len(procs))
	ratios := drng.FlatDirichlet(len(procs))
	for i, pidStr := range procs {
		if rand.Intn(999) < int(d.r.ProcParamDirichlet.ResetProbability*1000.0) {
			attrs[pidStr] = linuxsched.SchedAttr{
				Policy: linuxsched.Normal,
			}
		} else {
			// FIXME: we should obtain actual available NumCPU for the PID rather than runtime.NumCPU()
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
	}
	return attrs
}

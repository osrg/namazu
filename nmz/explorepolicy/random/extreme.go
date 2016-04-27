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
	"github.com/AkihiroSuda/go-linuxsched"
	log "github.com/cihub/seelog"
	"github.com/osrg/namazu/nmz/signal"
	"math/rand"
)

type extreme struct {
	r *Random
}

// implements procPolicyIntf
func (e *extreme) Action(event *signal.ProcSetEvent) (signal.Action, error) {
	procs, err := parseProcSetEvent(event)
	if err != nil {
		return nil, err
	}
	attrs := e.extremeSched(procs, e.r.PPPExtreme.Prioritized)
	for pidStr, attr := range attrs {
		log.Debugf("For PID=%s, setting Attr=%v", pidStr, attr)
	}
	return signal.NewProcSetSchedAction(event, attrs)
}

// Returns map of linuxsched.SchedAttr{} for procs
// due to JSON nature, we use string for PID representation.
func (e *extreme) extremeSched(procs []string, nprio int) map[string]linuxsched.SchedAttr {
	prios := make(map[int]bool)
	for i := 0; i < nprio; i++ {
		prios[int(rand.Int31n(int32(len(procs))))] = true
	}
	attrs := make(map[string]linuxsched.SchedAttr, len(procs))
	for i, pidStr := range procs {
		if prios[i] {
			attrs[pidStr] = linuxsched.SchedAttr{
				Policy:   linuxsched.RR,
				Priority: uint32(rand.Int31n(10)),
			}
		} else {
			attrs[pidStr] = linuxsched.SchedAttr{
				Policy: linuxsched.Batch,
			}
		}
	}
	return attrs
}

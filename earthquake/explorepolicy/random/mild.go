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

type mild struct {
	r *Random
}

// implements procPolicyIntf
func (e *mild) Action(event *signal.ProcSetEvent) (signal.Action, error) {
	procs, err := parseProcSetEvent(event)
	if err != nil {
		return nil, err
	}
	attrs := e.mildSched(procs, e.r.PPPMild.UseBatch)
	for pidStr, attr := range attrs {
		log.Debugf("For PID=%s, setting Attr=%v", pidStr, attr)
	}
	return signal.NewProcSetSchedAction(event, attrs)
}

// Returns map of linuxsched.SchedAttr{} for procs
// due to JSON nature, we use string for PID representation.
func (e *mild) mildSched(procs []string, useBatch bool) map[string]linuxsched.SchedAttr {
	attrs := make(map[string]linuxsched.SchedAttr, len(procs))
	for _, pidStr := range procs {
		policy := linuxsched.Normal
		if useBatch {
			policy = linuxsched.Batch
		}
		attrs[pidStr] = linuxsched.SchedAttr{
			Policy: policy,
			Nice:   int32(-20 + rand.Int31n(40)),
		}
	}
	return attrs
}

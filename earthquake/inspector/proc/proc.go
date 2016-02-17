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

package proc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AkihiroSuda/go-linuxsched"
	log "github.com/cihub/seelog"
	"github.com/osrg/earthquake/earthquake/inspector/transceiver"
	"github.com/osrg/earthquake/earthquake/signal"
	procutil "github.com/osrg/earthquake/earthquake/util/proc"
)

type ProcInspector struct {
	OrchestratorURL string
	EntityID        string
	RootPID         int
	WatchInterval   time.Duration
	trans           transceiver.Transceiver
}

func (this *ProcInspector) Start() error {
	log.Debugf("Initializing Process Inspector %#v", this)
	var err error

	this.trans, err = transceiver.NewTransceiver(this.OrchestratorURL, this.EntityID)
	if err != nil {
		return err
	}
	this.trans.Start()

	for {
		<-time.After(this.WatchInterval)
		procs, err := procutil.DescendantLWPs(this.RootPID)
		if err != nil {
			// this happens frequently, but does not matter.
			// e.g. "open /proc/11193/task/11193/children: no such file or directory"
			log.Warn(err)
			continue
		}
		if err = this.onWatch(procs); err != nil {
			log.Error(err)
		}
	}
	// NOTREACHED
}

func (this *ProcInspector) onWatch(procs []int) error {
	// due to JSON nature, use []string for procStrs
	procStrs := []string{}
	for _, proc := range procs {
		procStrs = append(procStrs, strconv.Itoa(proc))
	}
	event, err := signal.NewProcSetEvent(this.EntityID,
		procStrs, map[string]interface{}{})
	if err != nil {
		return err
	}
	actionCh, err := this.trans.SendEvent(event)
	if err != nil {
		return err
	}
	action := <-actionCh
	switch action.(type) {
	case *signal.ProcSetSchedAction:
		return this.onAction(action.(*signal.ProcSetSchedAction))
	default:
		return fmt.Errorf("unknown action %s. ignoring.", action)
	}
}

func (this *ProcInspector) onAction(action *signal.ProcSetSchedAction) error {
	// FIXME: this may not work with REST endpoint.
	// we need to convert interface{} to linuxsched.SchedAttr here

	attrs, ok := action.Option()["attrs"].(map[string]linuxsched.SchedAttr)
	if !ok {
		return fmt.Errorf("no attrs? this should be an implementation error. action=%#v", action)
	}

	for pidStr, attr := range attrs {
		// due to JSON nature, we use string for PID representation
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Warnf("Non PID string: %s", pidStr)
			continue
		}
		if warn := linuxsched.SetAttr(pid, attr); warn != nil {
			// this happens frequently, but does not matter.
			// so use log.Debugf rather than log.Warnf
			log.Debugf("could not apply %#v to %d: %s", attr, pid, warn)
		}
	}
	return nil
}

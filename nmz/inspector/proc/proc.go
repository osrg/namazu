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
	"syscall"
	"time"

	"github.com/AkihiroSuda/go-linuxsched"
	log "github.com/cihub/seelog"
	"github.com/mitchellh/mapstructure"

	"github.com/osrg/namazu/nmz/inspector/transceiver"
	"github.com/osrg/namazu/nmz/signal"
	procutil "github.com/osrg/namazu/nmz/util/proc"
)

type ProcInspector struct {
	OrchestratorURL string
	EntityID        string
	RootPID         int
	WatchInterval   time.Duration
	trans           transceiver.Transceiver
	// only for testing
	stopCh chan struct{}
}

func NewProcInspector(orchestratorURL, entityID string, rootPID int, watchInterval time.Duration) (*ProcInspector, error) {
	return &ProcInspector{
		OrchestratorURL: orchestratorURL,
		EntityID:        entityID,
		RootPID:         rootPID,
		WatchInterval:   watchInterval,
		stopCh:          make(chan struct{}),
	}, nil
}

func (this *ProcInspector) Serve(endCh <-chan struct{}) error {
	log.Debugf("Initializing Process Inspector %#v", this)
	var err error

	this.trans, err = transceiver.NewTransceiver(this.OrchestratorURL, this.EntityID)
	if err != nil {
		return err
	}
	this.trans.Start()

	for {
		select {
		case <-time.After(this.WatchInterval):
			procs, err := procutil.DescendantLWPs(this.RootPID)
			if err != nil {
				// this happens frequently, but does not matter.
				// e.g. "open /proc/11193/task/11193/children: no such file or directory"
				log.Warn(err)
				continue
			}
			if len(procs) == 0 {
				continue
			}
			if err = this.onWatch(procs); err != nil {
				log.Error(err)
			}
		case <-this.stopCh:
			log.Info("Shutting down..")
			return nil
		case <-endCh:
			log.Infof("Shutting down (via end channel)..")
			return nil
		}
	}
}

func (this *ProcInspector) Shutdown() {
	this.stopCh <- struct{}{}
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
	case *signal.NopAction:
		log.Debugf("nop action %s. ignoring.", action)
		return nil
	default:
		return fmt.Errorf("unknown action %s. ignoring.", action)
	}
}

// FIXME: due to JSON nature, we need complex type conversion, but it should be much simpler. this is the worst code ever
func parseAttrs(action *signal.ProcSetSchedAction) (map[string]linuxsched.SchedAttr, error) {
	// for local
	attrs, ok := action.Option()["attrs"].(map[string]linuxsched.SchedAttr)
	if ok {
		return attrs, nil
	}
	// for REST
	xattrs, ok := action.Option()["attrs"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no attrs? this should be an implementation error. action=%#v", action)
	}
	attrs = make(map[string]linuxsched.SchedAttr)
	for pidStr, xattr := range xattrs {
		mattr, ok := xattr.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("bad attr? this should be an implementation error. xattr=%#v, action=%#v", xattr, action)
		}
		attr := linuxsched.SchedAttr{}
		err := mapstructure.Decode(mattr, &attr)
		if err != nil {
			return nil, err
		}
		attrs[pidStr] = attr
	}
	return attrs, nil
}

func (this *ProcInspector) onAction(action *signal.ProcSetSchedAction) error {
	attrs, err := parseAttrs(action)
	if err != nil {
		return err
	}

	for pidStr, attr := range attrs {
		// due to JSON nature, we use string for PID representation
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Warnf("Non PID string: %s", pidStr)
			continue
		}
		if warn := linuxsched.SetAttr(pid, attr); warn != nil {
			if warn == syscall.EPERM {
				log.Errorf("could not apply %#v to %d: %v", attr, pid, warn)
			} else {
				// this happens frequently, but does not matter.
				// so use log.Debugf rather than log.Warnf
				log.Debugf("could not apply %#v to %d: %v", attr, pid, warn)
			}
		}
	}
	return nil
}

// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

/*
  Orchestrator manages the endpoint and the policy.

  Event:  inspector  --RPC-->  endpoint -> orchestrator -> policy
  Action: inspector <--RPC--   endpoint <- orchestrator <- policy

*/
package orchestrator

import (
	"runtime"
	"time"

	log "github.com/cihub/seelog"
	"github.com/osrg/namazu/nmz/endpoint"
	. "github.com/osrg/namazu/nmz/explorepolicy"
	. "github.com/osrg/namazu/nmz/signal"
	. "github.com/osrg/namazu/nmz/util/config"
	. "github.com/osrg/namazu/nmz/util/trace"

	"github.com/osrg/namazu/nmz/explorepolicy/dumb"
)

type Orchestrator struct {
	// arguments
	cfg          Config
	policy       ExplorePolicy
	dumbPolicy   ExplorePolicy // dumb policy is always required for enabling/disabling orchestration
	collectTrace bool
	// action sequence (can be so large)
	actionSequence []Action
	// communication channels
	endpointEventCh  chan Event
	endpointActionCh chan Action
	policyActionCh   chan Action
	controlCh            chan Control
	// orchestrator control channels
	stopEventRCh     chan struct{}
	stoppedEventRCh  chan struct{}
	stopActionRCh    chan struct{}
	stoppedActionRCh chan struct{}

	enabled bool
}

func NewOrchestrator(cfg Config, policy ExplorePolicy, collectTrace bool) *Orchestrator {
	orc := Orchestrator{
		cfg:            cfg,
		policy:         policy,
		dumbPolicy:     dumb.New(),
		collectTrace:   collectTrace,
		actionSequence: make([]Action, 0),
		// endpoint makes this
		endpointEventCh:  nil,
		endpointActionCh: make(chan Action),
		// policy makes this
		policyActionCh:   nil,
		stopEventRCh:     make(chan struct{}),
		stoppedEventRCh:  make(chan struct{}),
		stopActionRCh:    make(chan struct{}),
		stoppedActionRCh: make(chan struct{}),

		enabled: true,
	}
	return &orc
}

func (orc *Orchestrator) handleEvent(event Event) {
	if orc.enabled {
		log.Debugf("Orchestrator handling event %s", event)
		orc.policy.QueueEvent(event)
		log.Debugf("Orchestrator handled event %s", event)
	} else {
		log.Debugf("Orchestrator is disabled for now, ignoring %s", event)
		orc.dumbPolicy.QueueEvent(event)
		log.Debugf("ignored event %s", event)
	}
}

func (orc *Orchestrator) handleAction(action Action) {
	log.Debugf("Orchestrator handling action %s", action)
	var err error
	orcSideOnly := false
	orcSide, orcSideOk := action.(OrchestratorSideAction)
	action.SetTriggeredTime(time.Now())
	log.Debugf("action %s is executable on the orchestrator side: %t", action, orcSideOk)
	if orcSideOk {
		orcSideOnly = orcSide.OrchestratorSideOnly()
		log.Debugf("action %s is executable on only the orchestrator side: %t", action, orcSideOnly)
		err = orcSide.ExecuteOnOrchestrator()
		if err != nil {
			log.Errorf("ignoring an error occurred while ExecuteOnOrchestrator: %s", err)
		}
	}

	if !orcSideOnly {
		orc.endpointActionCh <- action
	}

	// make sequence for tracing
	if orc.collectTrace {
		orc.actionSequence = append(orc.actionSequence, action)
	}
	log.Debugf("Orchestrator handled action %s", action)
}

func (orc *Orchestrator) Start() {
	orc.endpointEventCh, orc.controlCh = endpoint.StartAll(orc.endpointActionCh, orc.cfg)
	orc.policyActionCh = orc.policy.ActionChan()
	go orc.eventRoutine()
	go orc.actionRoutine()
	go orc.controlRoutine()
}

func (orc *Orchestrator) eventRoutine() {
	defer close(orc.stoppedEventRCh)

	for {
		// NOTE: if NumGoroutine() increases rapidly, something is going wrong.
		log.Debugf("runtime.NumGoroutine()=%d", runtime.NumGoroutine())
		select {
		case event, ok := <-orc.endpointEventCh:
			if ok {
				// handleEvent() basically does: `policy.QueueEvent(event)`
				orc.handleEvent(event)
			} else {
				orc.endpointEventCh = nil
			}
		case <-orc.stopEventRCh:
			return
		}
		// connection lost to endpoints
		if orc.endpointEventCh == nil {
			close(orc.endpointActionCh)
		}
	}
}

func (orc *Orchestrator) actionRoutine() {
	defer close(orc.stoppedActionRCh)

	for {
		select {
		case action, ok := <-orc.policyActionCh:
			if ok {
				// handleAction() basically does: `endpointActionCh <- action`
				orc.handleAction(action)
			} else {
				orc.policyActionCh = nil
			}
		case <-orc.stopActionRCh:
			return
		}
	}
}

func (orc *Orchestrator) controlRoutine() {
	for {
		select {
		case control := <-orc.controlCh:
			switch control.Op {
			case ControlEnableOrchestrator:
				if orc.enabled {
					log.Warnf("orchestrator is already enabled")
				}
				orc.enabled = true
			case ControlDisableOrchestrator:
				if !orc.enabled {
					log.Warnf("orchestrator is already disabled")
				}
				orc.enabled = false
			default:
				log.Errorf("unkonown opcode of control: %d", control.Op)
			}
		}
	}
}

// Stops the orchestrator routine.
// Returns action trace if configured to do so.
func (orc *Orchestrator) Shutdown() *SingleTrace {
	log.Debugf("Shutting down orchestrator")
	close(orc.stopEventRCh)
	<-orc.stoppedEventRCh
	close(orc.stopActionRCh)
	<-orc.stoppedActionRCh
	newTrace := &SingleTrace{
		ActionSequence: orc.actionSequence,
	}
	log.Debugf("Action trace has %d actions", len(newTrace.ActionSequence))
	endpoint.ShutdownAll()
	log.Debugf("Shut down orchestrator")
	return newTrace
}

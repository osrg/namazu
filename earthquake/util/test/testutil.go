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

package test

import (
	"time"

	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/signal"
)

// used only for testing
type MockOrchestrator struct {
	// communication channels
	eventCh  chan Event
	actionCh chan Action
	// orchestrator control channels
	stopCh    chan struct{}
	stoppedCh chan struct{}
}

// used only for testing
func NewMockOrchestrator(eventCh chan Event, actionCh chan Action) *MockOrchestrator {
	orc := MockOrchestrator{
		// endpoint makes this
		eventCh:   eventCh,
		actionCh:  actionCh,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
	return &orc
}

func orchestratorSideOnlyAction(action Action) bool {
	orcSideOnly := false
	orcSide, orcSideOk := action.(OrchestratorSideAction)
	action.SetTriggeredTime(time.Now())
	if orcSideOk {
		orcSideOnly = orcSide.OrchestratorSideOnly()
	}
	return orcSideOnly
}

func (orc *MockOrchestrator) handleEvent(event Event) {
	action, err := event.DefaultAction()
	if err != nil {
		panic(log.Critical(err))
	}
	action.SetTriggeredTime(time.Now())
	if orchestratorSideOnlyAction(action) {
		panic(log.Critical("MockOrchestrator does not support OrchestratorSideOnly()"))
	}
	orc.actionCh <- action
}

func (orc *MockOrchestrator) routine() {
	defer close(orc.stoppedCh)
	for {
		select {
		case event, ok := <-orc.eventCh:
			if ok {
				orc.handleEvent(event)
			} else {
				orc.eventCh = nil
			}
		case <-orc.stopCh:
			return
		}
		// connection lost to endpoints
		if orc.eventCh == nil {
			close(orc.actionCh)
		}
	}
}

func (orc *MockOrchestrator) Start() {
	go orc.routine()
}

func (orc *MockOrchestrator) Shutdown() {
	close(orc.stopCh)
	<-orc.stoppedCh
}

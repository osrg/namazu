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
  Orchestrator manages inspectorhandlers, entities, and the policy.

  Event:  inspector  --RPC-->  inspectorhandler -> entity -> orchestrator -> policy
  Action: inspector <--RPC--   inspectorhandler <- entity <- orchestrator <- policy

*/
package orchestrator

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/entity"
	. "github.com/osrg/earthquake/earthquake/explorepolicy"
	. "github.com/osrg/earthquake/earthquake/inspectorhandler"
	. "github.com/osrg/earthquake/earthquake/signal"
	. "github.com/osrg/earthquake/earthquake/trace"
	. "github.com/osrg/earthquake/earthquake/util/config"
)

type Orchestrator struct {
	cfg            Config
	policy         ExplorePolicy
	collectTrace   bool
	actionSequence []Action
	running        bool
	endCh          chan interface{}
	newTraceCh     chan *SingleTrace
}

func NewOrchestrator(cfg Config, policy ExplorePolicy, collectTrace bool) *Orchestrator {
	oc := Orchestrator{
		cfg:            cfg,
		policy:         policy,
		collectTrace:   collectTrace,
		actionSequence: make([]Action, 0),
		running:        false,
		endCh:          make(chan interface{}),
		newTraceCh:     make(chan *SingleTrace),
	}
	return &oc
}

func (this *Orchestrator) handleAction(action Action) {
	var err error = nil
	ocSideOnly := false
	ocSide, ocSideOk := action.(OrchestratorSideAction)
	action.SetTriggeredTime(time.Now())
	log.Debugf("action %s is executable on the orchestrator side: %t", action, ocSideOk)
	if ocSideOk {
		ocSideOnly = ocSide.OrchestratorSideOnly()
		log.Debugf("action %s is executable on only the orchestrator side: %t", action, ocSideOnly)
		err = ocSide.ExecuteOnOrchestrator()
		if err != nil {
			log.Errorf("ignoring an error occurred while ExecuteOnOrchestrator: %s", err)
		}
	}

	if !ocSideOnly {
		// pass to the inspector handler.
		entity := GetTransitionEntity(action.EntityID())
		if entity == nil {
			err = fmt.Errorf("could find entity %s for %s", action.EntityID(), action)
			log.Errorf("ignoring an error: %s", err)
		} else {
			log.Debugf("Main[%s]->Handler: sending an action %s", entity.ID, action)
			entity.ActionFromMain <- action
			log.Debugf("Main[%s]->Handler: sent an action %s", entity.ID, action)
		}
	}

	// make sequence for tracing
	if this.collectTrace {
		this.actionSequence = append(this.actionSequence, action)
	}
}

func (this *Orchestrator) doDefaultAction(event Event) {
	action, err := event.DefaultAction()
	if err != nil {
		log.Errorf("ignoring an error: %s", err)
		return
	}
	this.handleAction(action)
}

func (this *Orchestrator) Start() {
	readyEntityCh := make(chan *TransitionEntity)
	StartAllInspectorHandler(readyEntityCh, this.cfg)
	policyNextActionChan := this.policy.GetNextActionChan()
	this.running = true
	log.Debugf("Main[running=%t]<-ExplorePolicy: receiving an action", this.running)
	for {
		select {
		case readyEntity := <-readyEntityCh:
			log.Debugf("Main[%s, running=%t]<-Handler: receiving an event", readyEntity.ID, this.running)
			event := <-readyEntity.EventToMain
			log.Debugf("Main[%s, running=%t]<-Handler: receiving an event %s", readyEntity.ID, this.running, event)
			if this.running && event.Deferred() {
				log.Debugf("Main[%s, running=%t]->ExplorePolicy: sending an event", readyEntity.ID, this.running)
				this.policy.QueueNextEvent(event)
				log.Debugf("Main[%s, running=%t]->ExplorePolicy: sent an event", readyEntity.ID, this.running)
			} else {
				// run script ended, accept event immediately without passing to the policy
				this.doDefaultAction(event)
			}
		case nextAction := <-policyNextActionChan:
			log.Debugf("Main[running=%t]<-ExplorePolicy: received an action %s", this.running, nextAction)
			this.handleAction(nextAction)
			log.Debugf("Main[running=%t]<-ExplorePolicy: receiving an action", this.running)
		case <-this.endCh:
			this.running = false
			newTrace := &SingleTrace{
				this.actionSequence,
			}
			this.newTraceCh <- newTrace
			// Do not really return from the function here, because the function must continue with running=false.
			// FIXME: when should we really return from the function after setting running=false?
		} // select
	} // for
	// NOTREACHED
}

func (this *Orchestrator) Shutdown() *SingleTrace {
	this.endCh <- true
	log.Debug("Receiving action trace")
	newTrace := <-this.newTraceCh
	log.Debugf("Received action trace (%d actions)", len(newTrace.ActionSequence))
	return newTrace
}

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

package local

import (
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/signal"
)

type LocalEndpoint struct {
	// only Orchestrator can access this
	orchestratorEventCh chan Event
	// only Orchestrator can access this
	orchestratorActionCh chan Action

	// only LocalTransceiver can access this
	InspectorEventCh chan Event
	// only LocalTransceiver can access this
	InspectorActionCh chan Action

	stopEventRCh     chan struct{}
	stoppedEventRCh  chan struct{}
	stopActionRCh    chan struct{}
	stoppedActionRCh chan struct{}
}

func (ep *LocalEndpoint) eventRoutine() {
	defer close(ep.stoppedEventRCh)
	// transceiver should do multiplexing
	for {
		select {
		case event, ok := <-ep.InspectorEventCh:
			if ok {
				log.Debugf("LOCAL EP handling event %s", event)
				ep.orchestratorEventCh <- event
				log.Debugf("LOCAL EP handled event %s", event)
			} else {
				ep.InspectorEventCh = nil
			}
		case <-ep.stopEventRCh:
			return
		}
		// connection lost
		if ep.InspectorEventCh == nil {
			close(ep.InspectorActionCh)
		}
		// connection lost
		if ep.orchestratorActionCh == nil {
			close(ep.orchestratorEventCh)
		}
	}
}

func (ep *LocalEndpoint) actionRoutine() {
	defer close(ep.stoppedActionRCh)
	// transceiver should do multiplexing
	for {
		select {
		case action, ok := <-ep.orchestratorActionCh:
			if ok {
				log.Debugf("LOCAL EP handling action %s", action)
				ep.InspectorActionCh <- action
				log.Debugf("LOCAL EP handled action %s", action)
			} else {
				ep.orchestratorActionCh = nil
			}
		case <-ep.stopActionRCh:
			return
		}
		// connection lost
		if ep.InspectorEventCh == nil {
			close(ep.InspectorActionCh)
		}
		// connection lost
		if ep.orchestratorActionCh == nil {
			close(ep.orchestratorEventCh)
		}
	}
}

func (ep *LocalEndpoint) Start(orchestratorActionCh chan Action) chan Event {
	ep.orchestratorActionCh = orchestratorActionCh
	ep.stopEventRCh = make(chan struct{})
	ep.stoppedEventRCh = make(chan struct{})
	ep.stopActionRCh = make(chan struct{})
	ep.stoppedActionRCh = make(chan struct{})
	go ep.eventRoutine()
	go ep.actionRoutine()
	return ep.orchestratorEventCh
}

func (ep *LocalEndpoint) Shutdown() {
	log.Debugf("Shutting down")
	close(ep.stopEventRCh)
	close(ep.stopActionRCh)
	<-ep.stoppedEventRCh
	<-ep.stoppedActionRCh
	log.Debugf("Shut down done")
}

func NewLocalEndpoint() LocalEndpoint {
	return LocalEndpoint{
		orchestratorEventCh: make(chan Event),
		// orchestrator makes (and can close) this
		orchestratorActionCh: nil,
		InspectorEventCh:     make(chan Event),
		InspectorActionCh:    make(chan Action),
	}
}

var SingletonLocalEndpoint = NewLocalEndpoint()

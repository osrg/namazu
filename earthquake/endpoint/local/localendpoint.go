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

	stop1Ch    chan struct{}
	stopped1Ch chan struct{}
	stop2Ch    chan struct{}
	stopped2Ch chan struct{}
}

func (ep *LocalEndpoint) routine1() {
	defer close(ep.stopped1Ch)
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
		case <-ep.stop1Ch:
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

func (ep *LocalEndpoint) routine2() {
	defer close(ep.stopped2Ch)
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
		case <-ep.stop2Ch:
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
	ep.stop1Ch = make(chan struct{})
	ep.stopped1Ch = make(chan struct{})
	ep.stop2Ch = make(chan struct{})
	ep.stopped2Ch = make(chan struct{})
	go ep.routine1()
	go ep.routine2()
	return ep.orchestratorEventCh
}

func (ep *LocalEndpoint) Shutdown() {
	log.Debugf("Shutting down")
	close(ep.stop1Ch)
	close(ep.stop2Ch)
	<-ep.stopped1Ch
	<-ep.stopped2Ch
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

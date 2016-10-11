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

// Package endpoint provides listeners for several RPC protocols
package endpoint

import (
	log "github.com/cihub/seelog"
	"github.com/osrg/namazu/nmz/endpoint/local"
	"github.com/osrg/namazu/nmz/endpoint/pb"
	"github.com/osrg/namazu/nmz/endpoint/rest"
	"github.com/osrg/namazu/nmz/signal"
	"github.com/osrg/namazu/nmz/util/config"
	"sync"
)

type endpointType int

const (
	endpointTypeLocal endpointType = iota
	endpointTypeREST
	endpointTypePB
)

var (
	muxEventCh  = make(chan signal.Event)
	muxActionCh chan signal.Action

	entityEndpointTypes   = make(map[string]endpointType)
	entityEndpointTypesMu = sync.RWMutex{}

	localEventCh  chan signal.Event
	localActionCh = make(chan signal.Action)
	restEventCh   chan signal.Event
	restActionCh  = make(chan signal.Action)
	pbEventCh     chan signal.Event
	pbActionCh    = make(chan signal.Action)

	stopActionRCh        chan struct{}
	stopLocalEventRCh    chan struct{}
	stopRESTEventRCh     chan struct{}
	stopPBEventRCh       chan struct{}
	stoppedActionRCh     chan struct{}
	stoppedLocalEventRCh chan struct{}
	stoppedRESTEventRCh  chan struct{}
	stoppedPBEventRCh    chan struct{}
)

// Starts all the endpoint handlers for multiplexed action channel actionCh.
// It returns an multiplexed event channel
func StartAll(actionCh chan signal.Action, cfg config.Config) (chan signal.Event, chan signal.Control) {
	muxActionCh = actionCh
	localEventCh = local.SingletonLocalEndpoint.Start(localActionCh)

	stopActionRCh = make(chan struct{})
	stopLocalEventRCh = make(chan struct{})
	stopRESTEventRCh = make(chan struct{})
	stopPBEventRCh = make(chan struct{})

	stoppedActionRCh = make(chan struct{})
	stoppedLocalEventRCh = make(chan struct{})
	stoppedRESTEventRCh = make(chan struct{})
	stoppedPBEventRCh = make(chan struct{})

	var controlCh chan signal.Control

	if cfg.IsSet("restPort") {
		restPort := cfg.GetInt("restPort")
		if restPort >= 0 {
			// zero is also legal (auto-assign)
			log.Infof("REST port: %d", restPort)
			restEventCh, controlCh = rest.SingletonRESTEndpoint.Start(restPort, restActionCh)
		} else {
			log.Warnf("ignoring restPort: %d", restPort)
		}
	}

	if cfg.IsSet("pbPort") {
		pbPort := cfg.GetInt("pbPort")
		if pbPort >= 0 {
			// zero is also legal (auto-assign)
			log.Infof("PB port: %d", pbPort)
			restEventCh = pb.SingletonPBEndpoint.Start(pbPort, restActionCh)
		} else {
			log.Warnf("ignoring pbPort: %d", pbPort)
		}
	}

	go actionRoutine()
	go localEventRoutine()
	go restEventRoutine()
	go pbEventRoutine()
	return muxEventCh, controlCh
}

func registerEntityEndpointType(entityID string, typ endpointType) {
	entityEndpointTypesMu.Lock()
	cur, ok := entityEndpointTypes[entityID]
	if ok {
		if cur != typ {
			log.Errorf("Entity ID conflict: new=%v, cur=%v", typ, cur)
			// this is very critical.. should we panic here?
		}
	} else {
		entityEndpointTypes[entityID] = typ
	}
	entityEndpointTypesMu.Unlock()
}

// Dispatch where action should be sent (localActionCh, restActionCh, pbActionCh..)
func dispatchAction(action signal.Action) {
	log.Debugf("EP handling action %s", action)
	entityID := action.EntityID()
	entityEndpointTypesMu.RLock()
	typ, ok := entityEndpointTypes[entityID]
	entityEndpointTypesMu.RUnlock()
	if ok {
		switch typ {
		case endpointTypeLocal:
			localActionCh <- action
		case endpointTypeREST:
			restActionCh <- action
		case endpointTypePB:
			pbActionCh <- action
		default:
			panic(log.Criticalf("Unknown endpoint type, cur=%s", entityID))
		}
	} else {
		log.Errorf("Unknown Entity ID:%s", entityID)
	}
	log.Debugf("EP handled action %s", action)
}

// Action sender (Orchestrator->Inspector)
func actionRoutine() {
	defer close(stoppedActionRCh)
	for {
		select {
		case action, ok := <-muxActionCh:
			if ok {
				dispatchAction(action)
			}
		case <-stopActionRCh:
			return
		}
	}
}

// only xxxEventRoutine() calls this
func onEvent(event signal.Event, typ endpointType) {
	log.Debugf("EP handling event %s", event)
	muxEventCh <- event
	registerEntityEndpointType(event.EntityID(), typ)
	log.Debugf("EP handled event %s", event)
}

// Local Event receiver (Inspector->Orchestrator)
func localEventRoutine() {
	defer close(stoppedLocalEventRCh)
	for {
		select {
		case event, ok := <-localEventCh:
			if ok {
				onEvent(event, endpointTypeLocal)
			}
		case <-stopLocalEventRCh:
			return
		}
	}
}

// REST Event receiver (Inspector->Orchestrator)
func restEventRoutine() {
	defer close(stoppedRESTEventRCh)
	for {
		select {
		case event, ok := <-restEventCh:
			if ok {
				onEvent(event, endpointTypeREST)
			}
		case <-stopRESTEventRCh:
			return
		}
	}
}

// PB Event receiver (Inspector->Orchestrator)
func pbEventRoutine() {
	defer close(stoppedPBEventRCh)
	for {
		select {
		case event, ok := <-pbEventCh:
			if ok {
				onEvent(event, endpointTypePB)
			}
		case <-stopPBEventRCh:
			return
		}
	}
}

func ShutdownAll() {
	log.Debugf("Shutting down")
	close(stopActionRCh)
	close(stopLocalEventRCh)
	close(stopRESTEventRCh)
	close(stopPBEventRCh)
	<-stoppedActionRCh
	<-stoppedLocalEventRCh
	<-stoppedRESTEventRCh
	<-stoppedPBEventRCh
	local.SingletonLocalEndpoint.Shutdown()
	// TODO: how to shutdown REST endpoint?
	log.Debugf("Shut down done")
}

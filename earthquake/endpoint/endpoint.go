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

package endpoint

import (
	"sync"

	log "github.com/cihub/seelog"
	"github.com/osrg/earthquake/earthquake/endpoint/local"
	"github.com/osrg/earthquake/earthquake/endpoint/rest"
	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
)

type endpointType int

const (
	endpointTypeLocal endpointType = iota
	endpointTypeREST
)

var (
	muxEventCh            = make(chan signal.Event)
	muxActionCh           chan signal.Action
	entityEndpointTypes   = make(map[string]endpointType)
	entityEndpointTypesMu = sync.RWMutex{}
	localEventCh          chan signal.Event
	localActionCh         = make(chan signal.Action)
	restEventCh           chan signal.Event
	restActionCh          = make(chan signal.Action)

	stopActionRCh        chan struct{}
	stopLocalEventRCh    chan struct{}
	stopRESTEventRCh     chan struct{}
	stoppedActionRCh     chan struct{}
	stoppedLocalEventRCh chan struct{}
	stoppedRESTEventRCh  chan struct{}
)

// Starts all the endpoint handlers for multiplexed action channel actionCh.
// It returns an multiplexed event channel
func StartAll(actionCh chan signal.Action, cfg config.Config) chan signal.Event {
	muxActionCh = actionCh
	localEventCh = local.SingletonLocalEndpoint.Start(localActionCh)

	stopActionRCh = make(chan struct{})
	stopLocalEventRCh = make(chan struct{})
	stopRESTEventRCh = make(chan struct{})
	stoppedActionRCh = make(chan struct{})
	stoppedLocalEventRCh = make(chan struct{})
	stoppedRESTEventRCh = make(chan struct{})

	if cfg.IsSet("pbPort") {
		pbPort := cfg.GetInt("pbPort")
		log.Warnf("ignoring pbPort (PB endpoint is disabled temporarily due to implementation issues in v0.2.0): %d", pbPort)
	}

	if cfg.IsSet("restPort") {
		restPort := cfg.GetInt("restPort")
		if restPort >= 0 {
			// zero is also legal (auto-assign)
			log.Infof("REST port: %d", restPort)
			restEventCh = rest.SingletonRESTEndpoint.Start(restPort, restActionCh)
		} else {
			log.Warnf("ignoring restPort: %d", restPort)
		}
	}

	go actionRoutine()
	go localEventRoutine()
	go restEventRoutine()
	return muxEventCh
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
	entityID := action.EntityID()

	entityEndpointTypesMu.RLock()
	cur, ok := entityEndpointTypes[entityID]
	entityEndpointTypesMu.RUnlock()
	if ok {
		switch cur {
		case endpointTypeLocal:
			localActionCh <- action
		case endpointTypeREST:
			restActionCh <- action
		default:
			log.Errorf("Unknown endpoint type, cur=%s", entityID)
		}
	} else {
		log.Errorf("Unknown Entity ID:%s", entityID)
	}
}

// Action sender (Orchestrator->Inspector)
func actionRoutine() {
	defer close(stoppedActionRCh)
	for {
		select {
		case action, ok := <-muxActionCh:
			if ok {
				log.Debugf("EP handling action %s", action)
				dispatchAction(action)
				log.Debugf("EP handled action %s", action)
			}
		case <-stopActionRCh:
			return
		}
	}
}

// Local Event receiver (Inspector->Orchestrator)
func localEventRoutine() {
	defer close(stoppedLocalEventRCh)
	for {
		select {
		case event, ok := <-localEventCh:
			if ok {
				log.Debugf("EP handling LOCAL event %s", event)
				muxEventCh <- event
				registerEntityEndpointType(event.EntityID(), endpointTypeLocal)
				log.Debugf("EP handled LOCAL event %s", event)
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
				log.Debugf("EP handling REST event %s", event)
				muxEventCh <- event
				registerEntityEndpointType(event.EntityID(), endpointTypeREST)
				log.Debugf("EP handled REST event %s", event)
			}
		case <-stopRESTEventRCh:
			return
		}
	}
}

func ShutdownAll() {
	log.Debugf("Shutting down")
	close(stopActionRCh)
	close(stopLocalEventRCh)
	close(stopRESTEventRCh)
	<-stoppedActionRCh
	<-stoppedLocalEventRCh
	<-stoppedRESTEventRCh
	local.SingletonLocalEndpoint.Shutdown()
	// TODO: how to shutdown REST endpoint?
	log.Debugf("Shut down done")
}

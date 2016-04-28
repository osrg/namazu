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

package transceiver

import (
	"fmt"
	"sync"

	log "github.com/cihub/seelog"
	localep "github.com/osrg/namazu/nmz/endpoint/local"
	. "github.com/osrg/namazu/nmz/signal"
)

type LocalTransceiver struct {
	m      map[string]chan Action // key: event id
	mMutex sync.Mutex
}

var SingletonLocalTransceiver = LocalTransceiver{
	m:      make(map[string]chan Action),
	mMutex: sync.Mutex{},
}

func (this *LocalTransceiver) SendEvent(event Event) (chan Action, error) {
	ch := make(chan Action)
	this.mMutex.Lock()
	// put ch to m BEFORE sending, otherwise race may occur
	this.m[event.ID()] = ch
	this.mMutex.Unlock()
	go func() {
		localep.SingletonLocalEndpoint.InspectorEventCh <- event
	}()
	return ch, nil
}

func (this *LocalTransceiver) onAction(action Action) error {
	event := action.Event()
	if event == nil {
		return fmt.Errorf("No event found for action %s", action)
	}
	this.mMutex.Lock()
	defer this.mMutex.Unlock()
	actionChan, ok := this.m[event.ID()]
	if !ok {
		return fmt.Errorf("No channel found for action %s (event id=%s)", action, event.ID())
	}
	delete(this.m, event.ID())
	// Namazu doesn't guarantee any determinism,
	// but can we make this more deterministic?
	go func() {
		actionChan <- action
	}()
	return nil
}

func (this *LocalTransceiver) routine() {
	onActionError := func(err error) {
		log.Error(err)
	}
	for {
		action := <-localep.SingletonLocalEndpoint.InspectorActionCh
		err := this.onAction(action)
		if err != nil {
			onActionError(err)
			continue
		}
	}
}

func (this *LocalTransceiver) Start() {
	// FIXME: should fail if already started
	go this.routine()
}

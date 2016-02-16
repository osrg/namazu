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
	log "github.com/cihub/seelog"
	"github.com/osrg/earthquake/earthquake/inspectorhandler"
	. "github.com/osrg/earthquake/earthquake/signal"
	"sync"
)

type LocalTransceiver struct {
	EntityID string
	m        map[string]chan Action // key: event id
	mMutex   sync.Mutex
}

func NewLocalTransceiver(entityID string) (Transceiver, error) {
	t := LocalTransceiver{
		EntityID: entityID,
		m:        make(map[string]chan Action),
		mMutex:   sync.Mutex{},
	}
	return &t, nil
}

func (this *LocalTransceiver) SendEvent(event Event) (chan Action, error) {
	if event.EntityID() != this.EntityID {
		return nil, fmt.Errorf("bad entity id for event %s (want %s)", event, this.EntityID)
	}
	ch := make(chan Action)
	this.mMutex.Lock()
	// put ch to m BEFORE calling SendEvent(), otherwise race may occur
	this.m[event.ID()] = ch
	this.mMutex.Unlock()
	go func() {
		inspectorhandler.GlobalLocalInspectorHandler.EventChan <- event
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
		action := <-inspectorhandler.GlobalLocalInspectorHandler.ActionChan
		err := this.onAction(action)
		if err != nil {
			onActionError(err)
			continue
		}
	}
}

func (this *LocalTransceiver) Start() {
	go this.routine()
}

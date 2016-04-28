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

package queue

import (
	"fmt"
	"sync"

	log "github.com/cihub/seelog"
	. "github.com/osrg/namazu/nmz/signal"
)

var (
	queues     = make(map[string]*ActionQueue, 0)
	queuesLock = sync.RWMutex{}
)

// concurrent-safe, deletable queue
type ActionQueue struct {
	EntityID         string
	peeking          bool
	peekingLock      sync.RWMutex
	actionsLock      sync.RWMutex
	actions          []Action
	actionsUpdatedCh chan bool
	killCh           chan bool
}

// locked
func (this *ActionQueue) Count() int {
	count := -1
	this.actionsLock.RLock()
	count = len(this.actions)
	this.actionsLock.RUnlock()
	return count
}

// locked
func (this *ActionQueue) Peeking() bool {
	peeking := false
	this.peekingLock.RLock()
	peeking = this.peeking
	this.peekingLock.RUnlock()
	return peeking
}

// locked, however old call loses(return nil); new one wins. (for concurrent http get)
func (this *ActionQueue) Peek() Action {
	log.Debugf("ActionQueue[%s]: Peeking", this.EntityID)
	if this.Peeking() {
		log.Warnf("ActionQueue[%s]: Concurrent peeking. Killing the old call.", this.EntityID)
		killed := make(chan bool)
		go func() {
			this.killCh <- true
			killed <- true
		}()
		<-killed
		log.Warnf("ActionQueue[%s]: Concurrent peeking. Killed the old call.", this.EntityID)
	}
	this.peekingLock.Lock()
	this.peeking = true
	this.peekingLock.Unlock()
	var action Action
	for {
		peeked := false
		this.actionsLock.RLock()
		if len(this.actions) > 0 {
			action = this.actions[0]
			peeked = true
		}
		this.actionsLock.RUnlock()
		if peeked {
			break
		}
		select {
		case <-this.actionsUpdatedCh:
			continue
		case <-this.killCh:
			log.Debugf("ActionQueue[%s]: killed, returning nil", this.EntityID)
			return nil
		}
	}
	log.Debugf("ActionQueue[%s]: Peeked action %s", this.EntityID, action)
	this.peekingLock.Lock()
	this.peeking = false
	this.peekingLock.Unlock()
	return action
}

// locked
func (this *ActionQueue) Put(action Action) {
	log.Debugf("ActionQueue[%s]: Putting %s", this.EntityID, action)
	this.actionsLock.Lock()
	oldLen := len(this.actions)
	this.actions = append(this.actions, action)
	this.actionsLock.Unlock()
	this.actionsUpdatedCh <- true
	log.Debugf("ActionQueue[%s]: Put(%d->%d) %s", this.EntityID, oldLen, oldLen+1, action)
}

// locked, idempotent
func (this *ActionQueue) Delete(actionUUID string) {
	log.Debugf("ActionQueue[%s]: Deleting %s", this.EntityID, actionUUID)
	this.actionsLock.Lock()
	defer this.actionsLock.Unlock()
	oldLen := len(this.actions)
	newActions := make([]Action, 0)
	for _, action := range this.actions {
		if action.ID() != actionUUID {
			newActions = append(newActions, action)
		}
	}
	newLen := len(newActions)
	deleted := oldLen - newLen
	if deleted > 1 {
		panic(log.Criticalf("this should not happen. deleted=%d", deleted))
	}
	if deleted > 0 {
		this.actions = newActions
	}
	log.Debugf("ActionQueue[%s]: Deleted(%d->%d) %s", this.EntityID, oldLen, newLen, actionUUID)
}

func RegisterNewQueue(entityID string) (*ActionQueue, error) {
	queuesLock.Lock()
	defer queuesLock.Unlock()
	old, oldOk := queues[entityID]
	if oldOk {
		return old, fmt.Errorf("entity exists %s(%#v)", entityID, old)
	}
	queue := ActionQueue{
		EntityID:         entityID,
		peeking:          false,
		peekingLock:      sync.RWMutex{},
		actionsLock:      sync.RWMutex{},
		actions:          make([]Action, 0),
		actionsUpdatedCh: make(chan bool),
		killCh:           make(chan bool),
	}
	queues[entityID] = &queue
	return &queue, nil
}

// intended for testing (`go test -count XXX` needs this)
func UnregisterQueue(entityID string) error {
	queuesLock.Lock()
	defer queuesLock.Unlock()
	_, oldOk := queues[entityID]
	if !oldOk {
		return fmt.Errorf("entity does not exists %s", entityID)
	}
	delete(queues, entityID)
	return nil
}

func GetQueue(entityID string) *ActionQueue {
	queuesLock.RLock()
	defer queuesLock.RUnlock()
	queue, ok := queues[entityID]
	if ok {
		return queue
	} else {
		return nil
	}
}

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

package restinspectorhandler

import (
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/entity"
	. "github.com/osrg/earthquake/earthquake/inspectorhandler/restinspectorhandler/queue"
	. "github.com/osrg/earthquake/earthquake/signal"
)

func registerNewEntity(entityID string) (*TransitionEntity, *ActionQueue, error) {
	log.Debugf("Registering entity: %s", entityID)
	entity := TransitionEntity{
		ID:             entityID,
		ActionFromMain: make(chan Action),
		EventToMain:    make(chan Event),
	}
	if err := RegisterTransitionEntity(&entity); err != nil {
		return nil, nil, err
	}
	queue, err := RegisterNewQueue(entityID)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		for {
			propagateActionFromMain(&entity, queue)
		}
	}()
	log.Debugf("Registered entity: %s", entityID)
	return &entity, queue, nil
}

func sendEventToMain(entity *TransitionEntity, event Event) {
	log.Debugf("Handler[%s]->Main: sending an event %s", entity.ID, event)
	sentToEntityCh := make(chan bool)
	sentToMainCh := make(chan bool)
	go func() {
		entity.EventToMain <- event
		sentToEntityCh <- true
	}()
	go func() {
		// StartAccept() registers mainReadyEntityCh
		mainReadyEntityCh <- entity
		sentToMainCh <- true
	}()
	<-sentToEntityCh
	<-sentToMainCh
	log.Debugf("Handler[%s]->Main: sent event %s", entity.ID, event)
}

func propagateActionFromMain(entity *TransitionEntity, queue *ActionQueue) {
	log.Debugf("Handler[%s]<-Main: receiving an action", entity.ID)
	action := <-entity.ActionFromMain
	log.Debugf("Handler[%s]<-Main: received an action %s", entity.ID, action)
	queue.Put(action)
}

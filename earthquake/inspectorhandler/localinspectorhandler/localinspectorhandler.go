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

package localinspectorhandler

import (
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/entity"
	. "github.com/osrg/earthquake/earthquake/signal"
)

type LocalInspectorHandler struct {
	EventChan  chan Event
	ActionChan chan Action
}

func (handler *LocalInspectorHandler) handleConn(readyEntityCh chan *TransitionEntity) {
	for {
		log.Debugf("Handler<-Inspector: receiving an event")
		event := <-handler.EventChan
		entityID := event.EntityID()
		log.Debugf("Handler[%s]<-Inspector: received an event %s", entityID, event)
		entity := GetTransitionEntity(entityID)
		if entity == nil {
			// FIXME: orchestrator requires this registration - can we remove this?
			entity = &TransitionEntity{
				ID:             entityID,
				ActionFromMain: make(chan Action),
				EventToMain:    make(chan Event),
			}
			err := RegisterTransitionEntity(entity)
			if err != nil {
				panic(log.Critical(err))
			}
			log.Debugf("Handler: initialized the entity structure", entityID)
		}

		log.Debugf("Handler[%s]->Main: sending an event %s", entityID, event)
		// FIXME: can we make this buffered?
		go func() {
			entity.EventToMain <- event
		}()
		readyEntityCh <- entity
		log.Debugf("Handler[%s]->Main: sent an event %s", entityID, event)

		log.Debugf("Handler[%s]<-Main: receiving an action", entityID)
		action := <-entity.ActionFromMain
		log.Debugf("Handler[%s]<-Main: received an action %s", entityID, action)

		log.Debugf("Handler[%s]->Inspector: sending an action %s", entityID, action)
		handler.ActionChan <- action
		log.Debugf("Handler[%s]->Inspector: sent an action %s", entityID, action)
	} // for
} // func

func (handler *LocalInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	go handler.handleConn(readyEntityCh)
}

func NewLocalInspectorHandler() *LocalInspectorHandler {
	return &LocalInspectorHandler{
		EventChan:  make(chan Event),
		ActionChan: make(chan Action),
	}
}

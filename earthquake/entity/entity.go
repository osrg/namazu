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

package entity

import (
	"fmt"

	. "github.com/osrg/earthquake/earthquake/signal"
)

var knownEntities = make(map[string]*TransitionEntity)

func RegisterTransitionEntity(entity *TransitionEntity) error {
	old, oldOk := knownEntities[entity.ID]
	if oldOk {
		return fmt.Errorf("overwriting an old entity %s(%#v)", entity.ID, old)
	}
	knownEntities[entity.ID] = entity
	return nil
}

func GetTransitionEntity(id string) *TransitionEntity {
	entity, ok := knownEntities[id]
	if ok {
		return entity
	}
	return nil
}

type TransitionEntity struct {
	// Entity ID string
	ID string

	// Event channel ( we don't have to use pointer for Event and Action, right? )
	EventToMain chan Event

	// Action channel
	ActionFromMain chan Action
}

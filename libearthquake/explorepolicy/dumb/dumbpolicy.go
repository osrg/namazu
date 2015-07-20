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

package dumb

import (
	. "../../equtils"
	. "../../historystorage"
)

type Dumb struct {
	nextActionChan chan *Action
}

func (d *Dumb) Init(storage HistoryStorage, param map[string]interface{}) {
}

func (d *Dumb) Name() string {
	return "dumb"
}

func (d *Dumb) GetNextActionChan() chan *Action {
	return d.nextActionChan
}

func (d *Dumb) QueueNextEvent(id string, ev *Event) {
	go func(e *Event) {
		act, err := e.MakeAcceptAction()
		if err != nil { panic(err) }
		d.nextActionChan <- act
	}(ev)
}

func DumbNew() *Dumb {
	nextActionChan := make(chan *Action)

	return &Dumb{
		nextActionChan,
	}
}

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
)

type Dumb struct {
	nextEventChan chan *Event
}

func (d *Dumb) Init() {
}

func (d *Dumb) Name() string {
	return "dumb"
}

func (d *Dumb) GetNextEventChan() chan *Event {
	return d.nextEventChan
}

func (d *Dumb) QueueNextEvent(ev *Event) {
	go func(e *Event) {
		d.nextEventChan <- ev
	}(ev)
}

func New() *Dumb {
	nextEventChan := make(chan *Event)

	return &Dumb{
		nextEventChan,
	}
}

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

package random

import (
	"math/rand"
	"sync"
	"time"

	. "../../equtils"
	. "../../historystorage"
)

type Random struct {
	nextEventChan chan *Event
	randGen       *rand.Rand
	queueMutex    *sync.Mutex
	eventQueue    []*Event
}

func (r *Random) Init(storage HistoryStorage) {
	go func() {
		for {
			// TODO: configurable
			time.Sleep(100 * time.Millisecond)

			r.queueMutex.Lock()
			qlen := len(r.eventQueue)
			if qlen == 0 {
				Log("no event is queued")
				r.queueMutex.Unlock()
				continue
			}

			idx := r.randGen.Int() % qlen
			next := r.eventQueue[idx]
			r.eventQueue = append(r.eventQueue[:idx], r.eventQueue[idx+1:]...)
			r.queueMutex.Unlock()

			r.nextEventChan <- next
		}
	}()
}

func (r *Random) Name() string {
	return "random"
}

func (r *Random) GetNextEventChan() chan *Event {
	return r.nextEventChan
}

func (r *Random) QueueNextEvent(ev *Event) {
	r.queueMutex.Lock()
	r.eventQueue = append(r.eventQueue, ev)
	r.queueMutex.Unlock()
}

func RandomNew() *Random {
	nextEventChan := make(chan *Event)
	eventQueue := make([]*Event, 0)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &Random{
		nextEventChan,
		r,
		mutex,
		eventQueue,
	}
}

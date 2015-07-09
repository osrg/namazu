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

type RandomParam struct {
	prioritize string
	interval   time.Duration // in millisecond
}

type Random struct {
	nextEventChan chan *Event
	randGen       *rand.Rand
	queueMutex    *sync.Mutex

	// todo: more than just two levels
	highEventQueue []*Event // high priority
	lowEventQueue  []*Event // low priority

	param *RandomParam
}

func constrRandomParam(rawParam map[string]interface{}) *RandomParam {
	var param RandomParam

	if _, ok := rawParam["prioritize"]; ok {
		param.prioritize = rawParam["prioritize"].(string)
	}

	if _, ok := rawParam["interval"]; ok {
		param.interval = time.Duration(int(rawParam["interval"].(float64)))
	} else {
		param.interval = time.Duration(100) // default: 100ms
	}

	return &param
}

func (r *Random) Init(storage HistoryStorage, param map[string]interface{}) {
	r.param = constrRandomParam(param)

	go func() {
		for {
			time.Sleep(r.param.interval * time.Millisecond)

			r.queueMutex.Lock()
			highLen := len(r.highEventQueue)
			lowLen := len(r.lowEventQueue)
			if highLen == 0 && lowLen == 0 {
				Log("no event is queued")
				r.queueMutex.Unlock()
				continue
			}

			var next *Event

			if highLen != 0 {
				idx := r.randGen.Int() % highLen
				next = r.highEventQueue[idx]
				r.highEventQueue = append(r.highEventQueue[:idx], r.highEventQueue[idx+1:]...)
			} else {
				idx := r.randGen.Int() % lowLen
				next = r.lowEventQueue[idx]
				r.lowEventQueue = append(r.lowEventQueue[:idx], r.lowEventQueue[idx+1:]...)
			}

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

func (r *Random) QueueNextEvent(procId string, ev *Event) {
	r.queueMutex.Lock()

	if r.param != nil && procId == r.param.prioritize {
		Log("**************** process %s alives, prioritizing\n", procId)
		r.highEventQueue = append(r.highEventQueue, ev)
	} else {
		r.lowEventQueue = append(r.lowEventQueue, ev)
	}
	r.queueMutex.Unlock()
}

func RandomNew() *Random {
	nextEventChan := make(chan *Event)
	highEventQueue := make([]*Event, 0)
	lowEventQueue := make([]*Event, 0)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &Random{
		nextEventChan,
		r,
		mutex,
		highEventQueue,
		lowEventQueue,
		nil,
	}
}

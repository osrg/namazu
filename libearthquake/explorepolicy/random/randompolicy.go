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

	timeBound bool
	maxBound  int // in millisecond
}

type Random struct {
	nextActionChan chan *Action
	randGen        *rand.Rand
	queueMutex     *sync.Mutex

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

	if _, ok := rawParam["timeBound"]; ok {
		param.timeBound = rawParam["timeBound"].(bool)
	}

	if _, ok := rawParam["maxBound"]; ok {
		param.maxBound = int(rawParam["maxBound"].(float64))
	} else {
		param.maxBound = 100 // default: 100ms
	}

	return &param
}

func (r *Random) Init(storage HistoryStorage, param map[string]interface{}) {
	r.param = constrRandomParam(param)

	if r.param.timeBound {
		return
	}

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

			act, err := next.MakeAcceptAction()
			if err != nil {
				panic(err)
			}
			r.nextActionChan <- act
		}
	}()
}

func (r *Random) Name() string {
	return "random"
}

func (r *Random) GetNextActionChan() chan *Action {
	return r.nextActionChan
}

func (r *Random) defaultQueueNextEvent(entityId string, ev *Event) {
	r.queueMutex.Lock()

	if r.param != nil && entityId == r.param.prioritize {
		Log("**************** entity %s alives, prioritizing\n", entityId)
		r.highEventQueue = append(r.highEventQueue, ev)
	} else {
		r.lowEventQueue = append(r.lowEventQueue, ev)
	}
	r.queueMutex.Unlock()
}

func (r *Random) timeBoundQueueNextEvent(entityId string, ev *Event) {
	go func(e *Event) {
		sleepMS := r.randGen.Int() % r.param.maxBound
		time.Sleep(time.Duration(sleepMS) * time.Millisecond)

		act, err := ev.MakeAcceptAction()
		if err != nil {
			panic(err)
		}

		r.nextActionChan <- act
	}(ev)
}

func (r *Random) QueueNextEvent(entityId string, ev *Event) {
	if r.param.timeBound {
		r.timeBoundQueueNextEvent(entityId, ev)
	} else {
		r.defaultQueueNextEvent(entityId, ev)
	}
}

func RandomNew() *Random {
	nextActionChan := make(chan *Action)
	highEventQueue := make([]*Event, 0)
	lowEventQueue := make([]*Event, 0)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &Random{
		nextActionChan,
		r,
		mutex,
		highEventQueue,
		lowEventQueue,
		nil,
	}
}

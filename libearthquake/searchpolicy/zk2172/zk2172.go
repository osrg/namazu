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

// BIG TODO: extract this as a loadable module

package zk2172

import (
	"math/rand"
	"sync"
	"time"

	. "../../equtils"
	. "../../historystorage"
)

type ZK2172Param struct {
	interval time.Duration // in millisecond
}

type zkEvent struct {
	id    string
	event *Event

	tick int
}

type ZK2172 struct {
	nextEventChan chan *Event
	randGen       *rand.Rand
	queueMutex    *sync.Mutex

	eventQueue []*zkEvent // high priority

	param *ZK2172Param
}

func constrParam(rawParam map[string]interface{}) *ZK2172Param {
	var param ZK2172Param

	if _, ok := rawParam["interval"]; ok {
		param.interval = time.Duration(int(rawParam["interval"].(float64)))
	} else {
		param.interval = time.Duration(100) // default: 100ms
	}

	return &param
}

func decreaseTick(queue []*zkEvent) {
	for _, ev := range queue {
		ev.tick--
	}
}

func (this *ZK2172) pickNextEvent() *Event {
	// TODO: use waiting time

	limit := 10
	var next *zkEvent
	for nrRetry := 0; nrRetry < limit; nrRetry++ {
		qlen := len(this.eventQueue)
		idx := this.randGen.Int() % qlen
		next = this.eventQueue[idx]

		if 0 < next.tick {
			continue
		}

		this.eventQueue = append(this.eventQueue[:idx], this.eventQueue[idx+1:]...)
		return next.event
	}

	return nil
}

func (this *ZK2172) Init(storage HistoryStorage, param map[string]interface{}) {
	this.param = constrParam(param)

	go func() {
		for {
			time.Sleep(this.param.interval * time.Millisecond)

			this.queueMutex.Lock()
			decreaseTick(this.eventQueue)

			qlen := len(this.eventQueue)
			if qlen == 0 {
				Log("no event is queued")
				this.queueMutex.Unlock()
				continue
			}

			nextEvent := this.pickNextEvent()
			this.queueMutex.Unlock()

			if nextEvent != nil {
				this.nextEventChan <- nextEvent
			}
		}
	}()
}

func (this *ZK2172) Name() string {
	return "ZK2172"
}

func (this *ZK2172) GetNextEventChan() chan *Event {
	return this.nextEventChan
}

func (this *ZK2172) QueueNextEvent(procId string, ev *Event) {
	newEv := &zkEvent{
		procId,
		ev,
		0,
	}

	if ev.EventParam == "deserializeSnapshot" && procId == "zksrv3" {
		Log("!!!!!!!!!!!!!!!!!! deserializeSnapshot arrives from %s", procId)
		newEv.tick = 1000
	}

	this.queueMutex.Lock()
	this.eventQueue = append(this.eventQueue, newEv)
	this.queueMutex.Unlock()
}

func ZK2172New() *ZK2172 {
	nextEventChan := make(chan *Event)
	eventQueue := make([]*zkEvent, 0)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &ZK2172{
		nextEventChan,
		r,
		mutex,
		eventQueue,
		nil,
	}
}

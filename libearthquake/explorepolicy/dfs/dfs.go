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

// dfs: an implementation of DFS explorer policy in earthquake
// As described in Junfeng Yang, et al. MODIST: Transparent Model Checking
// of Unmodified Distributed Systems. In Proc of NSDI '09. this policy is
// quite *useless*. It is just for experiment.

package dfs

import (
	log "github.com/cihub/seelog"
	"math/rand"
	"sync"
	"time"

	. "../../equtils"
	. "../../historystorage"
)

type DFSParam struct {
	interval time.Duration // in millisecond
}

type DFS struct {
	nextActionChan chan *Action
	randGen        *rand.Rand
	queueMutex     *sync.Mutex

	eventQueue []*Event // high priority

	dumb  bool
	param *DFSParam
}

func constrDFSParam(rawParam map[string]interface{}) *DFSParam {
	var param DFSParam

	if _, ok := rawParam["interval"]; ok {
		param.interval = time.Duration(int(rawParam["interval"].(float64)))
	} else {
		param.interval = time.Duration(100) // default: 100ms
	}

	return &param
}

func (this *DFS) Init(storage HistoryStorage, param map[string]interface{}) {
	this.param = constrDFSParam(param)

	prefix := make([]Event, 0)

	go func() {
		for {
			time.Sleep(this.param.interval * time.Millisecond)

			this.queueMutex.Lock()
			if len(this.eventQueue) == 0 {
				log.Debug("no event is queued")
				this.queueMutex.Unlock()
				continue
			}

			nextIdx := -1
			for i := 0; i < len(this.eventQueue); i++ {
				tmpprefix := append(prefix, *this.eventQueue[i])

				matched := storage.Search(tmpprefix)
				if len(matched) != 0 {
					nextIdx = i
					break
				}
			}

			if nextIdx == -1 {
				// no seen events in the past

				log.Debug("no matching histories")

				evQ := this.eventQueue
				this.eventQueue = []*Event{}
				this.dumb = true
				this.queueMutex.Unlock()

				go func() {
					for _, e := range evQ {
						act, err := e.MakeAcceptAction()
						if err != nil {
							panic(err)
						}
						this.nextActionChan <- act
					}
				}()

				// end this goroutine, below events are processed in a manner of dumb policy
				break
			}

			next := this.eventQueue[nextIdx]
			this.eventQueue = append(this.eventQueue[:nextIdx], this.eventQueue[nextIdx+1:]...)

			this.queueMutex.Unlock()

			prefix = append(prefix, *next)

			act, err := next.MakeAcceptAction()
			if err != nil {
				panic(err)
			}
			this.nextActionChan <- act
		}
	}()
}

func (this *DFS) Name() string {
	return "DFS"
}

func (this *DFS) GetNextActionChan() chan *Action {
	return this.nextActionChan
}

func (this *DFS) QueueNextEvent(entityId string, ev *Event) {
	this.queueMutex.Lock()

	if !this.dumb {
		this.eventQueue = append(this.eventQueue, ev)
	} else {
		go func() {
			act, err := ev.MakeAcceptAction()
			if err != nil {
				panic(err)
			}
			this.nextActionChan <- act
		}()
	}

	this.queueMutex.Unlock()
}

func DFSNew() *DFS {
	nextActionChan := make(chan *Action)
	eventQueue := make([]*Event, 0)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &DFS{
		nextActionChan,
		r,
		mutex,
		eventQueue,
		false,
		nil,
	}
}

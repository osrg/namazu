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

package bfs

import (
	"math/rand"
	"sync"
	"time"

	. "../../equtils"
	. "../../historystorage"
)

type BFSParam struct {
	interval time.Duration // in millisecond
}

type BFS struct {
	nextActionChan chan *Action
	randGen       *rand.Rand
	queueMutex    *sync.Mutex

	eventQueue []*Event // high priority

	dumb  bool
	param *BFSParam
}

func constrBFSParam(rawParam map[string]interface{}) *BFSParam {
	var param BFSParam

	if _, ok := rawParam["interval"]; ok {
		param.interval = time.Duration(int(rawParam["interval"].(float64)))
	} else {
		param.interval = time.Duration(100) // default: 100ms
	}

	return &param
}

func (this *BFS) Init(storage HistoryStorage, param map[string]interface{}) {
	this.param = constrBFSParam(param)

	prefix := make([]Event, 0)

	go func() {
		for {
			time.Sleep(this.param.interval * time.Millisecond)

			this.queueMutex.Lock()
			if len(this.eventQueue) == 0 {
				Log("no event is queued")
				this.queueMutex.Unlock()
				continue
			}

			nextIdx := -1
			for i := 0; i < len(this.eventQueue); i++ {
				// TODO: match every event in the queue, wait for more intervals, etc
				tmpprefix := append(prefix, *this.eventQueue[i])

				matched := storage.Search(tmpprefix)
				if len(matched) == 0 {
					nextIdx = i
					break
				}
			}

			if nextIdx == -1 {
				nextIdx = this.randGen.Int() % len(this.eventQueue)
			} else {
				Log("a completely new trace, go dumb mode")
				evQ := this.eventQueue
				this.eventQueue = []*Event{}
				this.dumb = true
				this.queueMutex.Unlock()

				go func() {
					for _, e := range evQ {
						act, err := e.MakeAcceptAction()
						if err != nil { panic(err) }
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
			if err != nil { panic(err) }			
			this.nextActionChan <- act
		}
	}()
}

func (this *BFS) Name() string {
	return "BFS"
}

func (this *BFS) GetNextActionChan() chan *Action {
	return this.nextActionChan
}

func (this *BFS) QueueNextEvent(procId string, ev *Event) {
	this.queueMutex.Lock()

	if !this.dumb {
		this.eventQueue = append(this.eventQueue, ev)
	} else {
		go func() {
			act, err := ev.MakeAcceptAction()
			if err != nil { panic(err) }
			this.nextActionChan <- act
		}()
	}

	this.queueMutex.Unlock()
}

func BFSNew() *BFS {
	nextActionChan := make(chan *Action)
	eventQueue := make([]*Event, 0)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &BFS{
		nextActionChan,
		r,
		mutex,
		eventQueue,
		false,
		nil,
	}
}

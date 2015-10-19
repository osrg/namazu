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
	"time"

	log "github.com/cihub/seelog"
	storage "github.com/osrg/earthquake/earthquake/historystorage"
	. "github.com/osrg/earthquake/earthquake/signal"
	queue "github.com/osrg/earthquake/earthquake/util/queue"
)

type Dumb struct {
	// channel
	nextActionChan chan Action

	// queue
	queue      queue.TimeBoundedQueue
	queueDeqCh chan queue.TimeBoundedQueueItem

	// parameter "interval"
	Interval time.Duration
}

// Initialize the policy
//
// storage can be nil
//
// parameters:
//
//  - interval(int): interval in millisecs (default: 0 msecs)
func (d *Dumb) Init(storage storage.HistoryStorage, param map[string]interface{}) {
	log.Debugf("Parameter: %#v", param)
	if v, ok := param["interval"].(float64); ok {
		d.Interval = time.Duration(int(v)) * time.Millisecond
	}
	log.Debugf("interval=%s", d.Interval)
	go d.dequeueEventRoutine()
}

// returns "dumb"
func (d *Dumb) Name() string {
	return "dumb"
}

func (d *Dumb) GetNextActionChan() chan Action {
	return d.nextActionChan
}

func (d *Dumb) dequeueEventRoutine() {
	for {
		qItem := <-d.queueDeqCh
		event := qItem.Value().(Event)
		action, err := event.DefaultAction()
		if err != nil {
			panic(log.Critical(err))
		}
		d.nextActionChan <- action
	}
}

func (d *Dumb) QueueNextEvent(event Event) {
	item, err := queue.NewBasicTBQueueItem(event, d.Interval, d.Interval)
	if err != nil {
		panic(log.Critical(err))
	}
	d.queue.Enqueue(item)
}

func New() *Dumb {
	q := queue.NewBasicTBQueue()
	return &Dumb{
		nextActionChan: make(chan Action),
		queue:          q,
		queueDeqCh:     q.GetDequeueChan(),
		Interval:       time.Duration(0),
	}
}

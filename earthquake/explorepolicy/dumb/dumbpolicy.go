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

// Package dumb provides the policy which does not control non-deternimism
package dumb

import (
	"time"

	log "github.com/cihub/seelog"
	"github.com/osrg/earthquake/earthquake/historystorage"
	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
	queue "github.com/osrg/earthquake/earthquake/util/queue"
)

type Dumb struct {
	// channel
	nextActionChan chan signal.Action

	// queue
	queue      queue.TimeBoundedQueue
	queueDeqCh chan queue.TimeBoundedQueueItem

	// parameter "interval"
	Interval time.Duration
}

func New() *Dumb {
	q := queue.NewBasicTBQueue()
	d := &Dumb{
		nextActionChan: make(chan signal.Action),
		queue:          q,
		queueDeqCh:     q.GetDequeueChan(),
		Interval:       time.Duration(0),
	}
	go d.dequeueEventRoutine()
	return d
}

const Name = "dumb"

// returns "dumb"
func (d *Dumb) Name() string {
	return Name
}

// parameters:
//  - interval(duration): interval (default: 0 msecs)
//
// should support dynamic reloading
func (d *Dumb) LoadConfig(cfg config.Config) error {
	log.Debugf("CONFIG: %s", cfg.AllSettings())
	paramInterval := "explorepolicyparam.interval"
	if cfg.IsSet(paramInterval) {
		d.Interval = cfg.GetDuration(paramInterval)
		log.Infof("Set interval=%s", d.Interval)
	} else {
		log.Infof("Using default interval=%s", d.Interval)
	}
	return nil
}

func (d *Dumb) SetHistoryStorage(storage historystorage.HistoryStorage) error {
	return nil
}

func (d *Dumb) ActionChan() chan signal.Action {
	return d.nextActionChan
}

func (d *Dumb) QueueEvent(event signal.Event) {
	item, err := queue.NewBasicTBQueueItem(event, d.Interval, d.Interval)
	if err != nil {
		panic(log.Critical(err))
	}
	d.queue.Enqueue(item)
}

func (d *Dumb) dequeueEventRoutine() {
	for {
		qItem := <-d.queueDeqCh
		event := qItem.Value().(signal.Event)
		action, err := event.DefaultAction()
		if err != nil {
			panic(log.Critical(err))
		}
		log.Debugf("DUMB: Determined action %s for event %s", action, event)
		d.nextActionChan <- action
	}
}

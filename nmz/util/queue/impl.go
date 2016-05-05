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

package queue

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/cihub/seelog"
	"github.com/eapache/channels"
)

// implements TimeBoundedQueueItem
type BasicTBQueueItem struct {
	value        interface{}
	enqueuedTime time.Time
	minDuration  time.Duration
	maxDuration  time.Duration
}

func NewBasicTBQueueItem(value interface{}, minDuration, maxDuration time.Duration) (*BasicTBQueueItem, error) {
	if minDuration > maxDuration {
		return nil, fmt.Errorf("minDuration %s > maxDuration%s", minDuration, maxDuration)
	}
	rand.Seed(time.Now().UnixNano())
	return &BasicTBQueueItem{
		value:        value,
		enqueuedTime: time.Unix(0, 0), // UNIX epoch (Jan 1, 1970 UTC)
		minDuration:  minDuration,
		maxDuration:  maxDuration,
	}, nil
}

func (this *BasicTBQueueItem) Value() interface{} {
	return this.value
}

func (this *BasicTBQueueItem) EnqueuedTime() time.Time {
	return this.enqueuedTime
}

func (this *BasicTBQueueItem) MinDuration() time.Duration {
	return this.minDuration
}

func (this *BasicTBQueueItem) MaxDuration() time.Duration {
	return this.maxDuration
}

// implements TimeBoundedQueue
type BasicTBQueue struct {
	dequeueChan        chan TimeBoundedQueueItem
	fixedDurationQueue *channels.InfiniteChannel
}

func NewBasicTBQueue() TimeBoundedQueue {
	q := &BasicTBQueue{
		dequeueChan:        make(chan TimeBoundedQueueItem),
		fixedDurationQueue: channels.NewInfiniteChannel(),
	}

	// TODO: gracefully shutdown this routine
	go func() {
		for {
			x := <-q.fixedDurationQueue.Out()
			item, ok := x.(TimeBoundedQueueItem)
			if ok {
				if item.MinDuration() != item.MaxDuration() {
					panic(fmt.Errorf("minDuration != maxDuration: %#v", item))
				}
				<-time.After(item.MaxDuration())
				q.dequeueChan <- item.(TimeBoundedQueueItem)
			}
		}
	}()

	return q
}

func determineDuration(minDuration, maxDuration time.Duration) time.Duration {
	x := int64(maxDuration - minDuration)
	y := int64(minDuration)
	r := time.Duration(rand.Int63n(x) + y)

	if r < minDuration {
		panic(fmt.Errorf("logic bug: %s < maxDuration=%s", r, minDuration))
	}
	if r > maxDuration {
		panic(fmt.Errorf("logic bug: %s > maxDuration=%s", r, maxDuration))
	}
	log.Debugf("determined duration %s (min:%s, max:%s)",
		r, minDuration, maxDuration)
	return r
}

func (this *BasicTBQueue) Enqueue(item_ TimeBoundedQueueItem) error {
	var item *BasicTBQueueItem
	item, ok := item_.(*BasicTBQueueItem)
	if !ok {
		return fmt.Errorf("bad item %s", item_)
	}
	item.enqueuedTime = time.Now()
	if item.minDuration == item.maxDuration {
		// we need this to ensure enqueuing order
		this.fixedDurationQueue.In() <- item
	} else {
		go func() {
			duration := determineDuration(item.minDuration, item.maxDuration)
			<-time.After(duration)
			this.dequeueChan <- item
		}()
	}
	return nil
}

func (this *BasicTBQueue) GetDequeueChan() chan TimeBoundedQueueItem {
	return this.dequeueChan
}

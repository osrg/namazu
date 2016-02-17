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

// Package queue provides TimeBoundedQueue
//
// the queue items are dequeued non-deterministically, but time bounds are set
package queue

import (
	"time"
)

// item for TimeBoundedQueue
type TimeBoundedQueueItem interface {
	// item value
	Value() interface{}

	// enqueued time
	EnqueuedTime() time.Time

	// min duration
	MinDuration() time.Duration

	// max duration (>= MinDuration)
	MaxDuration() time.Duration
}

// concurrent-safe, time-bounded queue.
//
// designed for ExplorePolicies.
type TimeBoundedQueue interface {
	// enqueue
	Enqueue(TimeBoundedQueueItem) error

	// get channel for dequeue
	GetDequeueChan() chan TimeBoundedQueueItem
}

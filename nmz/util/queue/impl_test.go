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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeBoundedQueueWithoutDurations(t *testing.T) {
	// for type assertion testing, declare var explicitly here
	var queue TimeBoundedQueue
	t.Logf("%s: Creating queue", time.Now())
	queue = NewBasicTBQueue()
	deqCh := queue.GetDequeueChan()
	t.Logf("%s: Created queue: %#v", time.Now(), queue)
	for i := 0; i < 3; i++ {
		item, err := NewBasicTBQueueItem(42+i, time.Duration(0), time.Duration(0))
		assert.NoError(t, err)
		t.Logf("%s: Enqueuing item: %#v", time.Now(), item)
		err = queue.Enqueue(item)
		assert.NoError(t, err)
		t.Logf("%s: Enqueued item: %#v", time.Now(), item)
	}

	t.Logf("%s: Dequeuing items", time.Now())
	for i := 0; i < 3; i++ {
		var deq TimeBoundedQueueItem
		deq = <-deqCh
		t.Logf("%s: Dequeued item: %#v", time.Now(), deq)
		assert.Equal(t, 42+i, deq.Value())
	}
}

func TestTimeBoundedQueueWithFixedDuration(t *testing.T) {
	queue := NewBasicTBQueue()
	deqCh := queue.GetDequeueChan()
	for i := 0; i < 3; i++ {
		item, err := NewBasicTBQueueItem(42+i, time.Duration(10*time.Millisecond), time.Duration(10*time.Millisecond))
		assert.NoError(t, err)
		err = queue.Enqueue(item)
		assert.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		var deq TimeBoundedQueueItem
		deq = <-deqCh
		assert.Equal(t, 42+i, deq.Value())
	}
}

func TestTimeBoundedQueueWithSeveralDurationsConcurrent(t *testing.T) {
	queue := NewBasicTBQueue()
	deqCh := queue.GetDequeueChan()

	durations := map[int][]time.Duration{
		42: {2000 * time.Millisecond, 7000 * time.Millisecond},
		43: {0 * time.Millisecond, 8000 * time.Millisecond},
		44: {5000 * time.Millisecond, 9000 * time.Millisecond},
	}

	for k, v := range durations {
		// CommonMistakes: Using goroutines on loop iterator variables
		// https://github.com/golang/go/wiki/CommonMistakes
		go func(k int, v []time.Duration) {
			t.Logf("%s: Enqueue %d, %s", time.Now(), k, v)
			item, err := NewBasicTBQueueItem(k, v[0], v[1])
			assert.NoError(t, err)
			err = queue.Enqueue(item)
			assert.NoError(t, err)
		}(k, v)
	}

	var wg sync.WaitGroup
	wg.Add(len(durations))
	delta := time.Duration(1000 * time.Millisecond)
	for i := 0; i < len(durations); i++ {
		go func() {
			defer wg.Done()
			deq := <-deqCh
			now := time.Now()
			took := now.Sub(deq.EnqueuedTime())
			t.Logf("%s: Took %s (%d, [%s, %s]@%s)",
				time.Now(),
				took,
				deq.Value(), deq.MinDuration(), deq.MaxDuration(), deq.EnqueuedTime())
			value := deq.Value().(int)
			assert.Equal(t, deq.MinDuration(), durations[value][0])
			assert.Equal(t, deq.MaxDuration(), durations[value][1])
			assert.True(t, took > deq.MinDuration())
			// can be delayed due to gorotuine scheduling, up to delta
			assert.True(t, took < deq.MaxDuration()+delta)
		}()
	}
	wg.Wait()
}

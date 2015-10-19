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
	"time"

	log "github.com/cihub/seelog"
	storage "github.com/osrg/earthquake/earthquake/historystorage"
	. "github.com/osrg/earthquake/earthquake/signal"
	queue "github.com/osrg/earthquake/earthquake/util/queue"
	"math/rand"
)

type Random struct {
	// channel
	nextActionChan chan Action

	// queue
	queue      queue.TimeBoundedQueue
	queueDeqCh chan queue.TimeBoundedQueueItem

	// parameter "minInterval"
	MinInterval time.Duration

	// parameter "maxInterval"
	MaxInterval time.Duration

	// parameter "prioritizedEntities"
	PrioritizedEntities map[string]bool

	// parameter "shellActionInterval"
	ShellActionInterval time.Duration

	// parameter "shellActionCommand"
	ShellActionCommand string

	// parameter "faultActionProbability”
	FaultActionProbability float64
}

// Initialize the policy
//
// storage can be nil
//
// parameters:
//
//  - minInterval(int): min interval in millisecs (default: 0 msecs)
//
//  - maxInterval(int): max interval (default == minInterval)
//
//  - prioritizedEntities([]string): prioritized entity string (default: empty)
//
//  - shellActionInterval(int): interval in millisecs for injecting ShellAction (default: 0)
//    NOTE: this can be 0 only if shellFaultCommand=""(empty string))
//
//  - shellActionCommand(string): command string for injecting ShellAction (default: empty string "")
//    NOTE: the command execution blocks.
//
//  - faultActionProbability(float64): probability (0.0-1.0) of PacketFaultAction/FilesystemFaultAction.
func (this *Random) Init(storage storage.HistoryStorage, param map[string]interface{}) {
	if v, ok := param["minInterval"].(float64); ok {
		this.MinInterval = time.Duration(int(v)) * time.Millisecond
	}
	if v, ok := param["maxInterval"].(float64); ok {
		this.MaxInterval = time.Duration(int(v)) * time.Millisecond
	} else {
		// set non-zero default value
		this.MaxInterval = this.MinInterval
	}
	log.Debugf("minInterval=%s, maxInterval=%s", this.MinInterval, this.MaxInterval)

	if v, ok := param["prioritizedEntities"].([]string); ok {
		for i := 0; i < len(v); i++ {
			this.PrioritizedEntities[v[i]] = true
		}
	}
	log.Debugf("prioritizedEntities=%s", this.PrioritizedEntities)

	if v, ok := param["shellActionInterval"].(float64); ok {
		this.ShellActionInterval = time.Duration(int(v)) * time.Millisecond
	}
	if v, ok := param["shellActionCommand"].(string); ok {
		this.ShellActionCommand = v
	}
	log.Debugf("shellActionInterval=%s, shellActionCommand=%s", this.ShellActionInterval, this.ShellActionCommand)
	if this.ShellActionInterval <= 0 && this.ShellActionCommand != "" {
		panic(log.Critical("shellFaultInterval must be positive value"))
	}

	if v, ok := param["faultActionProbability"].(float64); ok {
		this.FaultActionProbability = v
	}
	log.Debugf("faultActionProbability=%f", this.FaultActionProbability)
	if this.FaultActionProbability < 0.0 || this.FaultActionProbability > 1.0 {
		panic(log.Criticalf("bad faultActionProbability %f", this.FaultActionProbability))
	}

	go this.shellFaultInjectionRoutine()
	go this.dequeueEventRoutine()
}

// returns "random"
func (this *Random) Name() string {
	return "random"
}

func (this *Random) GetNextActionChan() chan Action {
	return this.nextActionChan
}

// put a ShellAction to nextActionChan
func (this *Random) shellFaultInjectionRoutine() {
	if this.ShellActionInterval == 0 {
		return
	}
	for {
		<-time.After(this.ShellActionInterval)
		// NOTE: you can also set arbitrary info (e.g., expected shutdown or unexpected kill)
		comments := map[string]interface{}{
			"comment": "injected by the random explorer",
		}
		action, err := NewShellAction(this.ShellActionCommand, comments)
		if err != nil {
			panic(log.Critical(err))
		}
		this.nextActionChan <- action
	}
}

// for dequeueRoutine()
func (this *Random) makeActionForEvent(event Event) (Action, error) {
	defaultAction, defaultActionErr := event.DefaultAction()
	faultAction, faultActionErr := event.DefaultFaultAction()
	if faultAction == nil {
		return defaultAction, defaultActionErr
	}
	if rand.Intn(999) < int(this.FaultActionProbability*1000.0) {
		log.Debugf("Injecting fault %s for %s", faultAction, event)
		return faultAction, faultActionErr
	} else {
		return defaultAction, defaultActionErr
	}
}

// dequeue event, determine corresponding action, and put the action to nextActionChan
func (this *Random) dequeueEventRoutine() {
	for {
		qItem := <-this.queueDeqCh
		event := qItem.Value().(Event)
		action, err := this.makeActionForEvent(event)
		if err != nil {
			panic(log.Critical(err))
		}
		this.nextActionChan <- action
	}
}

func (this *Random) QueueNextEvent(event Event) {
	minInterval := this.MinInterval
	maxInterval := this.MaxInterval
	_, prioritized := this.PrioritizedEntities[event.EntityID()]
	if prioritized {
		// FIXME: magic coefficient for prioritizing (decrease intervals)
		minInterval = time.Duration(float64(minInterval) * 0.8)
		maxInterval = time.Duration(float64(maxInterval) * 0.8)
	}
	item, err := queue.NewBasicTBQueueItem(event, minInterval, maxInterval)
	if err != nil {
		panic(log.Critical(err))
	}
	this.queue.Enqueue(item)
}

func New() *Random {
	nextActionChan := make(chan Action)
	q := queue.NewBasicTBQueue()
	return &Random{
		nextActionChan:         nextActionChan,
		queue:                  q,
		queueDeqCh:             q.GetDequeueChan(),
		MinInterval:            time.Duration(0),
		MaxInterval:            time.Duration(0),
		PrioritizedEntities:    make(map[string]bool, 0),
		ShellActionInterval:    time.Duration(0),
		ShellActionCommand:     "",
		FaultActionProbability: 0.0,
	}
}

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

// Package replayable provides the EXPERIMENTAL semi-deterministic replayable policy.
package replayable

import (
	"hash/fnv"
	"os"
	"time"

	log "github.com/cihub/seelog"
	"github.com/osrg/earthquake/earthquake/historystorage"
	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
)

type Replayable struct {
	// channel
	actionCh chan signal.Action

	// parameter "maxInterval"
	MaxInterval time.Duration

	// parameter "seed"
	Seed string
}

func New() *Replayable {
	log.Warnf("The replayable explorer is EXPERIMENTAL feature.")
	r := &Replayable{
		actionCh:    make(chan signal.Action),
		MaxInterval: time.Duration(0),
		Seed:        "",
	}
	return r
}

const Name = "replayable"

// returns "replayable"
func (r *Replayable) Name() string {
	return Name
}

// parameters:
//  - maxInterval(duration): max interval (default: 10 msecs)
//  - seed(string): seed for replaying (default: empty). can be overriden by EQ_REPLAY_SEED.
//
// should support dynamic reloading
func (r *Replayable) LoadConfig(cfg config.Config) error {
	log.Debugf("CONFIG: %s", cfg.AllSettings())
	paramMaxInterval := "explorepolicyparam.maxInterval"
	if cfg.IsSet(paramMaxInterval) {
		r.MaxInterval = cfg.GetDuration(paramMaxInterval)
		log.Infof("Set maxInterval=%s", r.MaxInterval)
	} else {
		r.MaxInterval = 10 * time.Millisecond
		log.Infof("Using default maxInterval=%s", r.MaxInterval)
	}

	paramSeed := "explorepolicyparam.seed"
	if cfg.IsSet(paramSeed) {
		r.Seed = cfg.GetString(paramSeed)
		log.Infof("Set seed=%s", r.Seed)
	} else {
		r.Seed = ""
		log.Infof("Using default seed=%s", r.Seed)
	}

	envSeed := "EQ_REPLAY_SEED"
	if v := os.Getenv(envSeed); v != "" {
		r.Seed = v
		log.Infof("Overriding seed=%s (%s)", r.Seed, envSeed)
	}

	return nil
}

func (d *Replayable) SetHistoryStorage(storage historystorage.HistoryStorage) error {
	return nil
}

func (d *Replayable) ActionChan() chan signal.Action {
	return d.actionCh
}

func (r *Replayable) determineInterval(event signal.Event) time.Duration {
	if r.MaxInterval == 0 {
		log.Warnf("MaxInterval is zero")
		return 0
	}
	hint := event.ReplayHint()
	h := fnv.New64a()
	h.Write([]byte(r.Seed))
	h.Write([]byte(hint))
	ui64 := h.Sum64()
	t := time.Duration(ui64 % uint64(r.MaxInterval))
	log.Debugf("REPLAYABLE: Determined interval %s for (seed=%s,hint=%s)",
		t, r.Seed, hint)
	return t
}

func (r *Replayable) QueueEvent(event signal.Event) {
	interval := r.determineInterval(event)
	action, err := event.DefaultAction()
	if err != nil {
		panic(log.Critical(err))
	}
	go func() {
		<-time.After(interval)
		r.actionCh <- action
	}()
}

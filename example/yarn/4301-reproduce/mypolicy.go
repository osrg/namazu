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

package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/cli"
	. "github.com/osrg/earthquake/earthquake/explorepolicy"
	. "github.com/osrg/earthquake/earthquake/historystorage"
	. "github.com/osrg/earthquake/earthquake/signal"
	config "github.com/osrg/earthquake/earthquake/util/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const MyPolicyName = "mypolicy-yarn-4301"
const FaultySleepDuration = 10 * time.Minute

type MyPolicy struct {
	nextActionChan chan Action
	sleep          time.Duration
}

func NewMyPolicy() ExplorePolicy {
	p := &MyPolicy{
		nextActionChan: make(chan Action),
		sleep:          0,
	}
	go p.signalHandler()
	return p
}

func (p *MyPolicy) Name() string {
	return MyPolicyName
}

func (p *MyPolicy) LoadConfig(cfg config.Config) error {
	return nil
}

func (p *MyPolicy) SetHistoryStorage(storage HistoryStorage) error {
	return nil
}

func (p *MyPolicy) signalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	for {
		sig := <-sigCh
		log.Infof("Caught %s, injecting sleep %s", sig, FaultySleepDuration)
		p.sleep = FaultySleepDuration
	}
}

func (p *MyPolicy) GetNextActionChan() chan Action {
	return p.nextActionChan
}

func formatSignalPair(event Event, action Action) string {
	eMap := event.JSONMap()
	eStr := fmt.Sprintf("%s(%v)", eMap["class"], eMap["option"])
	aMap := action.JSONMap()
	return fmt.Sprintf("%s for %s", aMap["class"], eStr)
}

func (p *MyPolicy) QueueNextEvent(event Event) {
	action, err := event.DefaultAction()
	if err != nil {
		panic(log.Critical(err))
	}
	go func() {
		if p.sleep > 0 {
			log.Infof("Sleping %s before %s", p.sleep,
				formatSignalPair(event, action))
			<-time.After(p.sleep)
			log.Infof("Slept %s. sending %s", p.sleep,
				formatSignalPair(event, action))
		} else {
			log.Infof("Sending %s", formatSignalPair(event, action))
		}
		p.nextActionChan <- action
	}()
}

func main() {
	log.Info("Earthquake for reproducing YARN 4301")
	RegisterPolicy(MyPolicyName, NewMyPolicy)
	os.Exit(CLIMain(os.Args))
}

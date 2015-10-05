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

// policy for etcd, based on random

package etcd

import (
	"encoding/base64"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	. "../../equtils"
	. "../../historystorage"
)

type killRatePerEntity struct {
	entityId string
	rate     int // 0 - 100
}

type shutdownRatePerEntity struct {
	entityId string
	rate     int // 0 - 100
}

type EtcdParam struct {
	prioritize string

	minBound int // in millisecond
	maxBound int // in millisecond

	killRates     []killRatePerEntity
	shutdownRates []shutdownRatePerEntity
}

type Etcd struct {
	nextActionChan chan *Action
	randGen        *rand.Rand
	queueMutex     *sync.Mutex

	param *EtcdParam
}

type etcdReq struct {
	method string
	path   string

	body []byte

	orig string
}

func parseEtcdReq(stringReq []byte) *etcdReq {
	// I don't know why but http.ReadResponse of go cannot parse etcd's request

	methodRe, merr := regexp.Compile("^(.+)")
	if merr != nil {
		Log("failed to parse regex: %s", merr)
		os.Exit(1)
	}

	method := methodRe.Find(stringReq)
	Log("method: %s", method)

	pathRe, perr := regexp.Compile("Path: (.+)")
	if perr != nil {
		Log("failed to parse regex: %s", perr)
		os.Exit(1)
	}

	path := pathRe.Find(stringReq)
	Log("path: %s", path)

	bodyIdx := strings.Index(string(stringReq), "\r\n\r\n")

	Log("bodyIdx: %d", bodyIdx)
	var body []byte
	if bodyIdx != -1 {
		body = stringReq[bodyIdx + 1:]
	}

	return &etcdReq{
		string(method),
		string(path),
		body,
		string(stringReq),
	}
}

func constrKillRatePerEntity(rawRates map[string]interface{}) []killRatePerEntity {
	rates := make([]killRatePerEntity, 0)

	for entityId, _rate := range rawRates {
		rate := int(_rate.(float64))
		newRate := killRatePerEntity{
			entityId: entityId,
			rate:     rate,
		}

		rates = append(rates, newRate)
	}

	return rates
}

func constrShutdownRatePerEntity(rawRates map[string]interface{}) []shutdownRatePerEntity {
	rates := make([]shutdownRatePerEntity, 0)

	for entityId, _rate := range rawRates {
		rate := _rate.(int)
		newRate := shutdownRatePerEntity{
			entityId: entityId,
			rate:     rate,
		}

		rates = append(rates, newRate)
	}

	return rates
}

func constrEtcdParam(rawParam map[string]interface{}) *EtcdParam {
	var param EtcdParam

	if _, ok := rawParam["maxBound"]; ok {
		param.maxBound = int(rawParam["maxBound"].(float64))
	} else {
		param.maxBound = 100 // default: 100ms
	}

	if _, ok := rawParam["minBound"]; ok {
		param.minBound = int(rawParam["minBound"].(float64))
	} else {
		param.minBound = 0 // default: 0ms
	}

	if _, ok := rawParam["killRatePerEntity"]; ok {
		param.killRates = constrKillRatePerEntity(rawParam["killRatePerEntity"].(map[string]interface{}))
	}

	if _, ok := rawParam["shutdownRatePerEntity"]; ok {
		param.shutdownRates = constrShutdownRatePerEntity(rawParam["shutdownRatePerEntity"].(map[string]interface{}))
	}

	return &param
}

func (policy *Etcd) shouldInjectFault(entityId string) bool {
	for _, rate := range policy.param.killRates {
		if rate.entityId != entityId {
			continue
		}

		return rate.rate < policy.randGen.Int()%100
	}

	return false
}

func (policy *Etcd) Init(storage HistoryStorage, param map[string]interface{}) {
	policy.param = constrEtcdParam(param)
}

func (policy *Etcd) Name() string {
	return "etcd"
}

func (policy *Etcd) GetNextActionChan() chan *Action {
	return policy.nextActionChan
}

func (policy *Etcd) QueueNextEvent(entityId string, ev *Event) {
	option := ev.EventParam["option"].(map[string]interface{})
	srcEntity := option["src_entity"].(string)
	dstEntity := option["dst_entity"].(string)
	msg := option["message"].(string)

	sep := strings.Index(srcEntity, ":")
	srcAddr := srcEntity[:sep]
	srcPortStr := srcEntity[sep+1:]
	srcPort, cerr := strconv.Atoi(srcPortStr)
	if cerr != nil {
		Log("converting src port (%s) failed: %s", srcPortStr, cerr)
		os.Exit(1)
	}

	sep = strings.Index(dstEntity, ":")
	dstAddr := dstEntity[:sep]
	dstPortStr := dstEntity[sep+1:]
	dstPort, cerr2 := strconv.Atoi(dstPortStr)
	if cerr2 != nil {
		Log("converting dst port (%s) failed: %s", dstPortStr, cerr2)
		os.Exit(1)
	}

	Log("src addr: %s, src port: %d, dst addr: %s, dst port: %d", srcAddr, srcPort, dstAddr, dstPort)
	if srcPort == 7001 { // response
		// TODO: correspondence between request and response

		go func() {
			act, err := ev.MakeAcceptAction()
			if err != nil {
				panic(err)
			}

			policy.nextActionChan <- act
		}()

		return
	}

	msgData, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		Log("fatal: decoding message (base64) failed: %s", err)
		os.Exit(1)
	}

	req := parseEtcdReq(msgData)
	Log("req: %v", req)

	go func(e *Event) {
		sleepMS := policy.randGen.Int() % policy.param.maxBound
		if sleepMS < policy.param.minBound {
			sleepMS = policy.param.minBound
		}

		time.Sleep(time.Duration(sleepMS) * time.Millisecond)

		act, err := ev.MakeAcceptAction()
		if err != nil {
			panic(err)
		}

		policy.nextActionChan <- act
	}(ev)
}

func EtcdNew() *Etcd {
	nextActionChan := make(chan *Action)
	mutex := new(sync.Mutex)
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return &Etcd{
		nextActionChan,
		r,
		mutex,
		nil,
	}
}

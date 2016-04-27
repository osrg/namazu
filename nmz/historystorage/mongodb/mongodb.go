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

package mongodb

import (
	naive "github.com/osrg/namazu/nmz/historystorage/naive"
	. "github.com/osrg/namazu/nmz/signal"
	. "github.com/osrg/namazu/nmz/util/trace"
	mgo "gopkg.in/mgo.v2"
	"time"
)

const (
	// TODO: make them configurable
	dialTo        = "mongodb://localhost/namazu"
	dbName        = "namazu"
	actionColName = "action"
	eventColName  = "event"
	traceColName  = "trace"
)

// type that implements interface HistoryStorage
type MongoDB struct {
	Naive   *naive.Naive
	dirPath string
	Session *mgo.Session
	DB      *mgo.Database
}

// TODO: get config

func New(dirPath string) *MongoDB {
	session, err := mgo.Dial(dialTo)
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	db := session.DB(dbName)

	return &MongoDB{
		Naive:   naive.New(dirPath),
		dirPath: dirPath,
		Session: session,
		DB:      db,
	}
}

func (this *MongoDB) CreateStorage() {
	this.Naive.CreateStorage()
}

func (this *MongoDB) Init() {
	this.Naive.Init()
}

func (this *MongoDB) Close() {
	this.Naive.Close()
	this.Session.Close()
}

func (this *MongoDB) Name() string {
	return "mongodb"
}

func (this *MongoDB) CreateNewWorkingDir() string {
	d := this.Naive.CreateNewWorkingDir()
	return d
}

func (this *MongoDB) RecordNewTrace(newTrace *SingleTrace) {
	this.Naive.RecordNewTrace(newTrace)

	traceDoc := map[string]interface{}{
		// FIXME: use something like this.Naive.GetCurrentTraceID()
		"trace_id": this.Naive.NrStoredHistories() - 1,
		"dir":      this.dirPath,
	}
	actionSequence := make([]map[string]interface{}, 0)
	for _, act := range newTrace.ActionSequence {
		actJSON := act.JSONMap()
		this.DB.C(actionColName).Insert(&actJSON)
		m := map[string]interface{}{
			// TODO: consider mongodb ObjectID?
			"uuid": actJSON["uuid"],
			"digest": map[string]interface{}{
				"class": actJSON["class"],
			},
		}
		evt := act.Event()
		if evt != nil {
			evtJSON := evt.JSONMap()
			this.DB.C(eventColName).Insert(&evtJSON)
			m["event_class"] = evtJSON["class"]
			m["event_option"] = evtJSON["option"]

		}
		actionSequence = append(actionSequence, m)
	}
	traceDoc["action_sequence"] = actionSequence
	this.DB.C(traceColName).Insert(&traceDoc)
}

func (this *MongoDB) RecordResult(successful bool, requiredTime time.Duration) error {
	return this.Naive.RecordResult(successful, requiredTime)
}

func (this *MongoDB) NrStoredHistories() int {
	nr := this.Naive.NrStoredHistories()
	return nr
}

func (this *MongoDB) GetStoredHistory(id int) (*SingleTrace, error) {
	trace, err := this.Naive.GetStoredHistory(id)
	return trace, err
}

func (this *MongoDB) IsSuccessful(id int) (bool, error) {
	succ, err := this.Naive.IsSuccessful(id)
	return succ, err
}

func (this *MongoDB) GetRequiredTime(id int) (time.Duration, error) {
	t, err := this.Naive.GetRequiredTime(id)
	return t, err
}

func (this *MongoDB) Search(prefix []Action) []int {
	slice := this.Naive.Search(prefix)
	return slice
}

func (this *MongoDB) SearchWithConverter(prefix []Action, converter func(actions []Action) []Action) []int {
	slice := this.Naive.SearchWithConverter(prefix, converter)
	return slice
}

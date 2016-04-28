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

package signal

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/AkihiroSuda/go-linuxsched"
	log "github.com/cihub/seelog"
	"github.com/kr/pretty"
)

func init() {
	gob.Register(FilesystemOp(""))
	gob.Register(map[string]interface{}{})
	gob.Register(linuxsched.SchedAttr{})
	gob.Register(map[string]linuxsched.SchedAttr{})
	gob.Register(time.Time{})
	gob.Register(BasicSignal{})
	gob.Register(&BasicEvent{}) // NOTE: use a pointer!
	gob.Register(&basicPBevent{})
	gob.Register(BasicAction{})
}

var knownSignalClasses = make(map[string]reflect.Type)

// Register an event class so that it can be serialized/deserialized
//
// name is a REST JSON class name
func RegisterSignalClass(name string, value interface{}) {
	log.Debugf("Registering a signal class \"%s\"", name)
	_, isEvent := value.(Event)
	_, isAction := value.(Action)
	if !(isEvent || isAction) {
		panic(log.Criticalf("%s is not an Event nor an action", value))
	}
	if _, registered := knownSignalClasses[name]; registered {
		panic(log.Criticalf("%s has been already registered", value))
	}

	t := reflect.TypeOf(value)
	knownSignalClasses[name] = t
	gob.Register(value)
}

// name is a REST JSON class name
func GetSignalClass(name string) *reflect.Type {
	if t, ok := knownSignalClasses[name]; ok {
		return &t
	}
	return nil
}

// Map-based Signal interface implementation
//
// Don't use pointer receiver for this type:
// https://github.com/golang/go/wiki/CodeReviewComments#receiver-type
type BasicSignal struct {
	M       map[string]interface{}
	Arrived time.Time
}

func (this *BasicSignal) InitSignal() {
	this.M = make(map[string]interface{})
	this.SetID("00000000-0000-0000-0000-000000000000")
	this.SetEntityID("_namazu_invalid_entity_id")
	this.SetType("_invalid_type")
	this.SetClass("_invalid_class")
	this.SetOption(map[string]interface{}{})
}

// for non-basic fields
func (this *BasicSignal) Get(k string) interface{} {
	v, ok := this.M[k]
	if !ok {
		return nil
	}
	return v
}

// for non-basic fields
func (this *BasicSignal) Set(k string, v interface{}) {
	this.M[k] = v
}

// implements Signal
func (this *BasicSignal) ID() string {
	return this.Get("uuid").(string)
}

func (this *BasicSignal) SetID(s string) {
	this.Set("uuid", s)
}

// implements Signal
func (this *BasicSignal) EntityID() string {
	return this.Get("entity").(string)
}

func (this *BasicSignal) SetEntityID(s string) {
	this.Set("entity", s)
}

// must be event or signal
func (this *BasicSignal) Type() string {
	return this.Get("type").(string)
}

// must be event or signal
func (this *BasicSignal) SetType(s string) error {
	if s != "event" && s != "action" {
		return fmt.Errorf("bad string %s", s)
	}
	this.Set("type", s)
	return nil
}

func (this *BasicSignal) Class() string {
	return this.Get("class").(string)
}

func (this *BasicSignal) SetClass(s string) {
	this.Set("class", s)
}

func (this *BasicSignal) Option() map[string]interface{} {
	return this.Get("option").(map[string]interface{})
}

func (this *BasicSignal) SetOption(m map[string]interface{}) {
	this.Set("option", m)
}

// implements Signal
func (this *BasicSignal) ArrivedTime() time.Time {
	return this.Arrived
}

func (this *BasicSignal) SetArrivedTime(t time.Time) {
	this.Arrived = t
}

// implements Signal
func (this *BasicSignal) JSONMap() map[string]interface{} {
	return this.M
}

func (this *BasicSignal) LoadJSONMap(m map[string]interface{}) error {
	this.M = make(map[string]interface{})
	for k, v := range m {
		this.M[k] = v
	}
	return nil
}

// implements Signal
func (this *BasicSignal) EqualsSignal(o Signal) bool {
	copied := BasicSignal{}
	copied.InitSignal()
	// from this, copy attributes that matters
	for k, v := range this.M {
		copied.Set(k, v)
	}
	// from o, copy attributes that can be ignored
	copied.SetID(o.ID())
	copied.SetArrivedTime(o.ArrivedTime())
	result := reflect.DeepEqual(o.JSONMap(), copied.JSONMap())
	return result
}

// implements Signal
func (this *BasicSignal) String() string {
	return pretty.Sprintf("Signal{%#v}", this.M)
}

func NewSignalFromJSONString(jsonString string, arrivedTime time.Time) (Signal, error) {
	var err error
	unmarshalled := make(map[string]interface{})
	if err = json.Unmarshal([]byte(jsonString), &unmarshalled); err != nil {
		return nil, err
	}
	name, ok := unmarshalled["class"].(string)
	if !ok {
		return nil, fmt.Errorf("bad json %#v", unmarshalled)
	}
	intf, err := instantiateSignalClass(name)
	if err != nil {
		return nil, err
	}
	signal, ok := intf.(ArrivalSignal)
	if !ok {
		return nil, fmt.Errorf("currently, non-basic signal cannot be instantiated with this method, intf=%#v", intf)
	}
	err = signal.LoadJSONMap(unmarshalled)
	if err != nil {
		return nil, err
	}
	signal.SetArrivedTime(arrivedTime)
	return signal.(Signal), nil
}

func instantiateSignalClass(name string) (interface{}, error) {
	var pClassType *reflect.Type
	if pClassType = GetSignalClass(name); pClassType == nil {
		return nil, fmt.Errorf("Unknown class \"%s\"", name)
	}

	// instantiation
	ref := reflect.New((*pClassType).Elem())
	intf := ref.Interface()
	_, isSignal := intf.(Signal)
	if !isSignal {
		return nil, fmt.Errorf("%s is not a signal", intf)
	}
	_, isEvent := intf.(Event)
	_, isAction := intf.(Action)
	if !(isEvent || isAction) {
		return nil, fmt.Errorf("%s is not an Event nor an Action", intf)
	}
	return intf, nil
}

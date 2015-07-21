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

package equtils

import (
	"fmt"
	"net"
	"reflect"
	"time"
	"github.com/satori/go.uuid"
)

// TODO: use viper, which enables aliasing for keeping compatibility
type EAParam map[string]interface{}

func (this EAParam) Equals(other EAParam) bool {
	return reflect.DeepEqual(this, other)
}

func NewEAParam() EAParam {
	eaParam := EAParam{}
	return eaParam
}

type Event_JavaSpecific_StackTraceElement struct {
	LineNumber int
	ClassName  string
	MethodName string
	FileName   string
}

type Event_JavaSpecific_Param struct {
	Name  string
	Value string
}

type Event_JavaSpecific struct {
	ThreadName string

	NrStackTraceElements int
	StackTraceElements   []Event_JavaSpecific_StackTraceElement

	NrParams int
	Params   []Event_JavaSpecific_Param
}

type Event struct {
	ArrivedTime time.Time

	ProcId string

	EventType  string // e.g., "FuncCall", "_JSON"
	EventParam EAParam

	JavaSpecific *Event_JavaSpecific
}

func (this Event) String() string {
	if this.EventType == "_JSON" {
		return fmt.Sprintf("JSONEvent{%s}", this.EventParam)
	} else {
		return fmt.Sprintf("Event{PID=%s, Type=%s, Param=%s}",
			this.ProcId, this.EventType, this.EventParam)
	}
}

type Action struct {
	ActionType  string // e.g., "Accept", "_JSON"
	ActionParam EAParam

	Evt *Event
}

func (this Action) String() string {
	if this.ActionType == "_JSON" {
		return fmt.Sprintf("JSONAction{%s}", this.ActionParam)
	} else {
		return fmt.Sprintf("Action{Type=%s, Param=%s, Event=%s}",
			this.ActionType, this.ActionParam, this.Evt)
	}
}

func (this *Event) MakeAcceptAction() (act *Action, err error) {
	if this.EventType != "_JSON" {
		// plain old events (e.g., "FuncCall")
		act = &Action{ActionType: "Accept", Evt: this}
	} else {
		// JSON events (for REST inspector handler)
		if ! this.EventParam["deferred"].(bool) {
			err = fmt.Errorf("Cannot accept an event of which \"deferred\" is false")
			return
		}
		act = &Action {
			ActionType: "_JSON",
			ActionParam: EAParam{
				// TODO: wrap me
				// please refer to JSON schema file for this format
				"type": "action",
				"class": "AcceptDeferredEventAction",
				"process": this.ProcId,
				"uuid": uuid.NewV4().String(),
				"option": map[string]interface{} {
					"event_uuid": this.EventParam["uuid"].(string),
				},
			},
			Evt: this,
		}
	}
	return
}

type SingleTrace struct {
	ActionSequence []Action // NOTE: Action holds the corresponding Evt
}

type TransitionEntity struct {
	Id   string
	Conn net.Conn

	EventToMain chan *Event
	ActionFromMain chan  *Action
}

func compareJavaSpecificFields(a, b *Event) bool {
	// skip thread name and stack trace currently

	if a.JavaSpecific.NrParams != b.JavaSpecific.NrParams {
		return false
	}

	for i, aParam := range a.JavaSpecific.Params {
		bParam := &b.JavaSpecific.Params[i]

		if aParam.Name != bParam.Name {
			return false
		}

		if aParam.Value != bParam.Value {
			return false
		}
	}

	return true
}

func AreEventsEqual(a, b *Event) bool {
	if a.ProcId != b.ProcId {
		return false
	}

	if a.EventType != b.EventType {
		return false
	}

	if ! a.EventParam.Equals(b.EventParam) {
		return false
	}

	if a.JavaSpecific != nil && b.JavaSpecific != nil{
		return compareJavaSpecificFields(a, b)
	}

	return true
}

func AreEventsSliceEqual(a, b []Event) bool{
	aLen := len(a)
	bLen := len(b)
	if aLen != bLen {
		return false
	}

	for i := 0; i < aLen; i++ {
		if !AreEventsEqual(&a[i], &b[i]) {
			return false
		}
	}

	return true
}

func AreActionsSliceEqual(a, b []Action) bool{
	aLen := len(a)
	bLen := len(b)
	if aLen != bLen {
		return false
	}

	for i := 0; i < aLen; i++ {
		if !AreEventsEqual(a[i].Evt, b[i].Evt) {
			return false
		}
	}

	return true
}

func AreTracesEqual(a, b *SingleTrace) bool {
	return  AreActionsSliceEqual(a.ActionSequence, b.ActionSequence)
}

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
	"github.com/satori/go.uuid"
	"net"
	"reflect"
	"time"
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

	EntityId string

	EventId    string // used by MongoDB and so on. expected to compliant with RFC 4122 UUID string format
	EventType  string // e.g., "FuncCall", "_JSON"
	EventParam EAParam

	JavaSpecific *Event_JavaSpecific
}

func (this Event) Validate() error {
	if this.EventType == "_JSON" {
		// TODO: check JSON schema
	} else {
		if this.EventType == "FuncCall" {
			// nop
		} else if this.EventType == "FuncReturn" {
			// nop
		} else {
			return fmt.Errorf("Unknown EventType %s", this.EventType)
		}
	}
	if this.EventId == "" {
		return fmt.Errorf("EventId not set")
	}
	return nil
}

func (this Event) String() string {
	if this.EventType == "_JSON" {
		return fmt.Sprintf("JSONEvent{%s}", this.EventParam)
	} else {
		return fmt.Sprintf("Event{PID=%s, Type=%s, Param=%s}",
			this.EntityId, this.EventType, this.EventParam)
	}
}

func (this *Event) ToJSONMap() map[string]interface{} {
	if this.EventType == "_JSON" {
		return this.EventParam
	}
	m := map[string]interface{}{
		// please refer to JSON schema file for this format
		"type":     "event",
		"class":    "",
		"deferred": true,
		"entity":   this.EntityId,
		"uuid":     this.EventId,
		"option":   map[string]interface{}{},
	}
	if this.EventType == "FuncCall" {
		m["class"] = "FunctionCallEvent"
		m["option"].(map[string]interface{})["func_name"] = this.EventParam["name"]
	} else if this.EventType == "FuncReturn" {
		m["class"] = "FunctionReturnEvent"
		m["option"].(map[string]interface{})["func_name"] = this.EventParam["name"]
	} else {
		Panic("invalid type of event: %d", this.EventType)
	}
	if this.JavaSpecific != nil {
		m["option"].(map[string]interface{})["thread_name"] = this.JavaSpecific.ThreadName
		stackTrace := make([]map[string]interface{}, 0)
		for _, stackTraceElement := range this.JavaSpecific.StackTraceElements {
			element := map[string]interface{}{
				"line_number": stackTraceElement.LineNumber,
				"class_name":  stackTraceElement.ClassName,
				"method_name": stackTraceElement.MethodName,
				"file_name":   stackTraceElement.FileName,
			}
			stackTrace = append(stackTrace, element)
		}
		m["option"].(map[string]interface{})["stack"] = stackTrace
		for _, param := range this.JavaSpecific.Params {
			m["option"].(map[string]interface{})[param.Name] = param.Value
		}
	}
	return m
}

func EventFromJSONMap(m map[string]interface{}, arrivedTime time.Time, entityId string) (ev Event, err error) {
	ev = Event{
		ArrivedTime: arrivedTime,
		EntityId:    entityId,
		EventId:     m["uuid"].(string),
		EventType:   "_JSON",
		EventParam:  m,
	}
	err = ev.Validate()
	return
}

func (this *FaultEvent) ToJSONMap() map[string]interface{} {
	return map[string]interface{}{
		"entity":  this.EntityId,
		"uuid":    this.uuid,
		"EventId": this.uuid,
	}
}

type FaultEvent struct {
	EntityId      string
	TriggeredTime time.Time
	uuid          string
}

type Action struct {
	ActionId    string // used by MongoDB and so on. expected to compliant with RFC 4122 UUID string format
	ActionType  string // e.g., "Accept", "_JSON"
	ActionParam EAParam

	Evt      *Event
	FaultEvt *FaultEvent
}

func (this Action) Validate() error {
	if this.ActionType == "_JSON" {
		// TODO: check JSON schema
	} else {
		if this.ActionType == "Accept" {
			if this.Evt == nil {
				return fmt.Errorf("No event tied")
			}
		} else {
			return fmt.Errorf("Unknown ActionType %s", this.ActionType)
		}
	}
	if this.ActionId == "" {
		return fmt.Errorf("ActionId not set")
	}
	return nil
}

func (this Action) String() string {
	if this.ActionType == "_JSON" {
		return fmt.Sprintf("JSONAction{%s}", this.ActionParam)
	} else {
		return fmt.Sprintf("Action{Type=%s, Param=%s, Event=%s}",
			this.ActionType, this.ActionParam, this.Evt)
	}
}

func (this *Action) ToJSONMap() map[string]interface{} {
	if this.ActionType == "_JSON" {
		return this.ActionParam
	} else if this.ActionType == "Accept" {
		// NOTE: this.Evt: PB Event, jsonEvent: JSON Event
		jsonEvent, err := EventFromJSONMap(this.Evt.ToJSONMap(), this.Evt.ArrivedTime, this.Evt.EntityId)
		if err != nil {
			Panic("%s", err)
		}
		jsonAction, err := jsonEvent.MakeAcceptAction()
		if err != nil {
			Panic("%s", err)
		}
		return jsonAction.ToJSONMap()
	} else if this.ActionType == "Kill" {
		jsonEvent, err := EventFromJSONMap(this.FaultEvt.ToJSONMap(), this.FaultEvt.TriggeredTime, this.FaultEvt.EntityId)
		if err != nil {
			Panic("%s", err)
		}
		jsonAction, err := jsonEvent.MakeAcceptAction()
		if err != nil {
			Panic("%s", err)
		}
		return jsonAction.ToJSONMap()
	} else {
		// FIXME: return an error
		Panic("Unknown type %s", this.ActionType)
		return nil
	}
}

func (this *Event) MakeAcceptAction() (act *Action, err error) {
	actionId := uuid.NewV4().String()
	if this.EventType != "_JSON" {
		// plain old events (e.g., "FuncCall")
		act = &Action{ActionId: actionId, ActionType: "Accept", Evt: this}
	} else {
		// JSON events (for REST inspector handler)
		if !this.EventParam["deferred"].(bool) {
			err = fmt.Errorf("Cannot accept an event of which \"deferred\" is false")
			return
		}
		act = &Action{
			ActionId:   actionId,
			ActionType: "_JSON",
			ActionParam: EAParam{
				// TODO: wrap me
				// please refer to JSON schema file for this format
				"type":   "action",
				"class":  "AcceptDeferredEventAction",
				"entity": this.EntityId,
				"uuid":   actionId,
				"option": map[string]interface{}{
					"event_uuid": this.EventParam["uuid"].(string),
				},
			},
			Evt: this,
		}
	}
	err = act.Validate()
	return
}

type SingleTrace struct {
	ActionSequence []Action // NOTE: Action holds the corresponding Evt
}

type TransitionEntity struct {
	Id   string
	Conn net.Conn

	EventToMain    chan *Event
	ActionFromMain chan *Action
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
	if a.EntityId != b.EntityId {
		return false
	}

	if a.EventType != b.EventType {
		return false
	}

	if !a.EventParam.Equals(b.EventParam) {
		return false
	}

	if a.JavaSpecific != nil && b.JavaSpecific != nil {
		return compareJavaSpecificFields(a, b)
	}

	// we don't have to care about EventId, right?

	return true
}

func AreEventsSliceEqual(a, b []Event) bool {
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

func AreActionsSliceEqual(a, b []Action) bool {
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
	return AreActionsSliceEqual(a.ActionSequence, b.ActionSequence)
}

func MakeFaultInjectionAction(entityId string) *Action {
	actionId := uuid.NewV4().String()
	// plain old events (e.g., "FuncCall")
	return &Action{
		ActionId:   actionId,
		ActionType: "Kill",
		ActionParam: EAParam{
			"type":   "action",
			"class":  "KillDeferredEventAction",
			"entity": entityId,
			"uuid":   actionId,
			"option": map[string]interface{}{
				"event_uuid": actionId,
			},
		},
		Evt: nil,
		FaultEvt: &FaultEvent{
			uuid:          actionId,
			EntityId:      entityId,
			TriggeredTime: time.Now(),
		},
	}
	// TODO: json events
}

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
	"net"
	"time"
)

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

	EventType  string
	EventParam string

	JavaSpecific *Event_JavaSpecific
}

type SingleTrace struct {
	EventSequence []Event
}

type TransitionEntity struct {
	Id   string
	Conn net.Conn

	EventToMain chan *Event
	GotoNext  chan interface{}
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

	if a.EventParam != b.EventParam {
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

func AreTracesEqual(a, b *SingleTrace) bool {
	if len(a.EventSequence) != len(b.EventSequence) {
		return false
	}

	for i := 0; i < len(a.EventSequence); i++ {
		if !AreEventsEqual(&a.EventSequence[i], &b.EventSequence[i]) {
			return false
		}
	}

	return true
}

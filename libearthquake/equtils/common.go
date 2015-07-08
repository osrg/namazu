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


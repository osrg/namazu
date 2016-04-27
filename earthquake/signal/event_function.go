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
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	pb "github.com/osrg/namazu/nmz/util/pb"
	"github.com/satori/go.uuid"
)

type FunctionEventType int

const (
	NilFunctionEventType FunctionEventType = iota
	FunctionCall
	FunctionReturn
)

// implements Event
//
// not implemented yet
type CFunctionEvent struct {
	basicPBevent
	FunctionName      string
	FunctionEventType FunctionEventType
}

// implements Event
type JavaFunctionEvent struct {
	basicPBevent
	FunctionName       string
	FunctionEventType  FunctionEventType
	ThreadName         string
	StackTraceElements []Event_JavaSpecific_StackTraceElement
	Params             []Event_JavaSpecific_Param
}

// for Java
type Event_JavaSpecific_Param struct {
	Name  string
	Value string
}

// for Java
type Event_JavaSpecific_StackTraceElement struct {
	LineNumber int
	ClassName  string
	MethodName string
	FileName   string
}

func NewJavaFunctionEventFromPB(msg pb.InspectorMsgReq, arrivedTime time.Time) (Event, error) {
	event := JavaFunctionEvent{}
	event.InitSignal()
	// set basic fields
	event.SetArrivedTime(arrivedTime)
	event.SetID(uuid.NewV4().String())
	event.SetEntityID(msg.GetEntityId())
	event.SetType("action")
	event.SetClass("JavaFunctionEvent")
	event.SetDeferred(true)
	opt := map[string]interface{}{}

	// set pbReq and pbRes
	event.pbReq = msg
	ackCode := pb.InspectorMsgRsp_ACK
	event.pbRes = pb.InspectorMsgRsp{
		Res:   &ackCode,
		MsgId: proto.Int32(msg.GetMsgId()),
	}

	if msg.GetEvent().GetType() == pb.InspectorMsgReq_Event_FUNC_CALL {
		event.FunctionEventType = FunctionCall
		opt["function_event_type"] = "call"
		event.FunctionName = msg.GetEvent().GetFuncCall().GetName()
	} else if msg.GetEvent().GetType() == pb.InspectorMsgReq_Event_FUNC_RETURN {
		event.FunctionEventType = FunctionReturn
		opt["function_event_type"] = "return"
		event.FunctionName = msg.GetEvent().GetFuncReturn().GetName()
	} else {
		return &event, fmt.Errorf("Bad event type %#v", msg.Event.Type)
	}
	opt["function_name"] = event.FunctionName

	if msg.GetHasJavaSpecificFields() == 1 {
		event.ThreadName = msg.GetJavaSpecificFields().GetThreadName()
		opt["thread_name"] = event.ThreadName
		optParams := map[string]interface{}{}
		for _, pbParam := range msg.GetJavaSpecificFields().GetParams() {
			param := Event_JavaSpecific_Param{
				Name:  pbParam.GetName(),
				Value: pbParam.GetValue(),
			}
			optParams[pbParam.GetName()] = pbParam.GetValue()
			event.Params = append(event.Params, param)
		}
		opt["params"] = optParams

		// TODO: put stack trace to opt
		for _, pbElem := range msg.GetJavaSpecificFields().GetStackTraceElements() {
			element := Event_JavaSpecific_StackTraceElement{
				LineNumber: int(pbElem.GetLineNumber()),
				ClassName:  pbElem.GetClassName(),
				MethodName: pbElem.GetMethodName(),
				FileName:   pbElem.GetFileName(),
			}
			event.StackTraceElements = append(event.StackTraceElements, element)
		}
	}

	event.SetOption(opt)
	return &event, nil
}

// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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
	// "code.google.com/p/goprotobuf/proto"
	// "encoding/binary"
	. "./equtils"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync/atomic"
)

type executionGlobalFlags struct {
	direct bool
	// if direct is true, inspected applications talk with orchestrator directly
	// there's no guest agents and VMs
}

type ProcessState int

const (
	PROCESS_STATE_CONNECTED ProcessState = 1 // before initiation, FIXME: it is not suitable for non direct mode
	PROCESS_STATE_READY     ProcessState = 2 // initiation completed
	PROCESS_STATE_DEAD      ProcessState = 3 // dead, connection closed
)

type machine struct {
	idx int
	id  string

	virtIOInPath  string
	virtIOOutPath string

	QMPTCPPort int
	QMPConn    net.Conn

	virtIOIn  net.Conn // unix domain socket
	virtIOOut net.Conn // unix domain socket

	eventReqRecv chan *I2GMsgReq
	eventRspSend chan *I2GMsgRsp

	processes []*process
}

type process struct {
	exe *execution

	id  string
	idx int

	machineID string // only for non direct mode

	conn  net.Conn
	state ProcessState

	startExecution chan interface{}
	endExecution   chan bool
	endCompletion  chan interface{}

	eventReqToMain chan *I2GMsgReq_Event
	gotoNext       chan interface{}

	eventReqRecv chan *I2GMsgReq
	eventRspSend chan *I2GMsgRsp

	directRecvEndCompletion chan interface{} // only used in direct mode (FIXME)
	directSendEndCompletion chan interface{}
}

type executionUnitState struct {
	processId string

	eventType  string
	eventParam interface{}
}

type executionUnitAction struct {
	actionType string
	param      interface{}
}

type executionUnit struct {
	states []executionUnitState
	action executionUnitAction
}

type executionSequence []executionUnit

type execution struct {
	globalFlags executionGlobalFlags

	machines  []*machine
	processes []*process
	sequence  executionSequence

	initiationCompletion chan int
	eventArrive          chan int

	directListen net.Listener // only used in direct mode
}

const (
	SearchModeInfoPath = "SearchModeInfo"
)

type SearchModeInfo struct {
	NrCollectedTraces int
}

var exe *execution

func parseExecutionFile(path string, simulation bool) *execution {
	jsonBuf, err := WholeRead(path)
	if err != nil {
		Log("reading execution file: %s failed (%s)", path, err)
		return nil
	}

	exe := &execution{}
	var root map[string]interface{}

	err = json.Unmarshal(jsonBuf, &root)
	if err != nil {
		Log("unmarsharing execution file: %s failed (%s)", path, err)
		return nil
	}

	globalFlags := root["globalFlags"].(map[string]interface{})
	exe.globalFlags.direct = int(globalFlags["direct"].(float64)) == 1

	if !exe.globalFlags.direct {
		machines := root["machines"].([]interface{})
		for _, _machine := range machines {
			id := _machine.(map[string]interface{})["id"].(string)
			virtIOInPath := _machine.(map[string]interface{})["virtIOIn"].(string)
			virtIOOutPath := _machine.(map[string]interface{})["virtIOOut"].(string)
			QMPTCPPort := int(_machine.(map[string]interface{})["QMPTCPPort"].(float64))

			newMachine := &machine{
				id:            id,
				virtIOInPath:  virtIOInPath,
				virtIOOutPath: virtIOOutPath,
				QMPTCPPort:    QMPTCPPort,
			}

			initiateQMP(newMachine)

			exe.machines = append(exe.machines, newMachine)

			Log("new machine, ID: %s, virtIO in path: %s, virtIO out path: %s", newMachine.id, newMachine.virtIOInPath, newMachine.virtIOOutPath)
		}
	}

	processes := root["processes"].([]interface{})

	for _, _process := range processes {
		id := _process.(map[string]interface{})["id"].(string)

		for _, existProcess := range exe.processes {
			if id == existProcess.id {
				Panic("Process ID: %s is duplicated", id)
			}
		}

		newProcess := process{
			id: id,
		}

		if !exe.globalFlags.direct {
			newProcess.machineID = _process.(map[string]interface{})["machineID"].(string)

			appended := false
			for _, m := range exe.machines {
				if m.id != newProcess.machineID {
					continue
				}

				m.processes = append(m.processes, &newProcess)
				appended = true
				break
			}

			if !appended {
				Log("invalid machine ID: %s of process: %s", newProcess.machineID, newProcess.id)
			}
		}

		exe.processes = append(exe.processes, &newProcess)
	}

	if !simulation {
		// search mode doesn't require executionSequence
		return exe
	}

	sequence := root["executionSequence"].([]interface{})

	for _, _unit := range sequence {
		unit := _unit.(map[string]interface{})
		states := unit["states"].([]interface{})

		newUnitStates := make([]executionUnitState, 0)

		for _, _state := range states {
			state := _state.(map[string]interface{})
			processId := state["processId"].(string)

			for _, s := range newUnitStates {
				if s.processId == processId {
					Panic("Process ID: %s is duplicated in single executionUnitState", processId)
				}
			}

			event := state["event"].(map[string]interface{})
			typ := event["eventType"].(string)
			param := event["eventParam"].(interface{})

			newUnitState := executionUnitState{
				processId:  processId,
				eventType:  typ,
				eventParam: param,
			}

			newUnitStates = append(newUnitStates, newUnitState)
		}

		action := unit["action"].(map[string]interface{})
		newUnitAction := executionUnitAction{
			actionType: action["type"].(string),
			param:      action["param"],
		}

		newExecutionUnit := executionUnit{
			states: newUnitStates,
			action: newUnitAction,
		}

		exe.sequence = append(exe.sequence, newExecutionUnit)
	}

	return exe
}

func handleProcess(n *process) {
	if n.exe.globalFlags.direct {
		req := <-n.eventReqRecv

		if *req.Type != I2GMsgReq_INITIATION {
			Log("invalid message during initiation (process index: %d)", n.idx)
			os.Exit(1)
		}

		initiationReq := req.Initiation
		Log("Process %s is initiating...", *initiationReq.ProcessId)
		n.id = *initiationReq.ProcessId

		result := I2GMsgRsp_ACK
		req_msg_id := *req.MsgId
		rsp := &I2GMsgRsp{
			Res:   &result,
			MsgId: &req_msg_id,
		}

		n.eventRspSend <- rsp

		Log("initiation of process %s succeed (handler part)", n.id)
		exe.initiationCompletion <- n.idx
	}

	<-n.startExecution
	Log("start execution of process %s", n.id)

	recvCh := make(chan bool)

	req := (*I2GMsgReq)(nil)

	go func() {
		for {
			req = <-n.eventReqRecv
			Log("received event from main goroutine: %s", n.id)
			recvCh <- true
		}
	}()

	for {

		select {
		case <-recvCh:
			if *req.Type != I2GMsgReq_EVENT {
				Log("invalid message from process %s, type: %d", n.id, *req.Type)
				os.Exit(1)
			}

			Log("event message received from process %s", n.id)
			var gaMsgId int32
			if !n.exe.globalFlags.direct {
				gaMsgId = *req.GaMsgId
				Log("guest agent message id: %d", gaMsgId)
			}

			go func(r *I2GMsgReq) {
				n.eventReqToMain <- r.Event
			}(req)

			exe.eventArrive <- n.idx
			if *req.Event.Type != I2GMsgReq_Event_EXIT {
				<-n.gotoNext

				result := I2GMsgRsp_ACK
				req_msg_id := *req.MsgId
				rsp := &I2GMsgRsp{
					Res:   &result,
					MsgId: &req_msg_id,
				}
				if !n.exe.globalFlags.direct {
					rsp.GaMsgId = &gaMsgId
				}

				n.eventRspSend <- rsp
				Log("replied to the event message from process %s", n.id)
			}
		case needToNotifyEnd := <-n.endExecution:
			Log("goroutine of process: %s received end message from main goroutine", n.id)
			Log("needToNotifyEnd: %s", needToNotifyEnd)

			if needToNotifyEnd {
				result := I2GMsgRsp_END
				rsp := &I2GMsgRsp{
					Res: &result,
				}

				n.eventRspSend <- rsp

				Log("sent end message to process: %s", n.id)

				if n.exe.globalFlags.direct {
					Log("waiting for completion of recv/send routines (process: %s)", n.id)
					<-n.directRecvEndCompletion
					<-n.directSendEndCompletion
					Log("completed recv/send routines (process: %s)", n.id)
				}
			}

			Log("sending endCompletion")
			n.endCompletion <- true

			Log("inspection end (process: %s)", n.id)
			return
		}

	}
}

func matchStateAndEvent(state *executionUnitState, ev *I2GMsgReq_Event) bool {
	if state.eventType == "funcCall" {
		if *ev.Type == I2GMsgReq_Event_FUNC_CALL {
			ev_funcCall_name := *ev.FuncCall.Name
			state_funcCall := state.eventParam.(map[string]interface{})
			if state_funcCall["name"].(string) == ev_funcCall_name {
				Log("matched, funcCall")
				return true
			} else {
				Log("not matched, funcCall")
			}

			return false
		}

		return false
	}

	return false
}

func runExecution() {
	for _, n := range exe.processes {
		Log("starting execution of process %s", n.id)
		n.startExecution <- true
	}
	Log("all processes started")

	for uNumber, u := range exe.sequence {
		Log("starting unit %d", uNumber)

		pendingProcesses := make([]*process, 0)

		states := u.states
		nrWaitingStates := len(states)
		Log("nrWaitingStates: %d", nrWaitingStates)
		for nrWaitingStates != 0 {
			nIdx := <-exe.eventArrive
			n := exe.processes[nIdx]

			found := false

			state := (*executionUnitState)(nil)
			for _, s := range states {
				if s.processId != n.id {
					continue
				}

				state = &s
				found = true
				break
			}

			eventReq := <-n.eventReqToMain
			if !found {
				n.gotoNext <- true
				continue
			}

			if !matchStateAndEvent(state, eventReq) {
				n.gotoNext <- true
				continue
			}

			pendingProcesses = append(pendingProcesses, n)
			nrWaitingStates--
		}

		doAction(u.action)

		for _, n := range pendingProcesses {
			n.gotoNext <- true
		}

		Log("unit %d end", uNumber)
	}

	Log("sending end notification to goroutines")
	for _, n := range exe.processes {
		n.endExecution <- true
	}

	Log("gathering end completion notification from goroutines")
	nrEnded := int32(0)
	fin := make(chan interface{})
	for _, n := range exe.processes {
		go func(p *process) {
			<-p.endCompletion
			atomic.AddInt32(&nrEnded, 1)
			if int(nrEnded) == len(exe.processes) {
				fin <- true
			}
		}(n)
	}
	<-fin
}

func machineInitiation(m *machine, fin chan int) {
	Log("initiating with machine %s", m.id)

	req := &G2OMsgReq{}
	rerr := RecvMsg(m.virtIOIn, req)
	if rerr != nil {
		Log("receiving initiation message from machine %s failed", m.id)
		os.Exit(1)
	}

	if *req.Op != G2OMsgReq_INITIATION {
		Log("invalid opcode: %d", *req.Op)
		os.Exit(1)
	}

	if *req.Initiation.Id != m.id {
		Log("invalid machine ID: %s (correct: %s)", req.Initiation.Id, m.id)
		os.Exit(1)
	}

	res := G2OMsgRsp_SUCCEED
	rsp := &G2OMsgRsp{
		Res: &res,
	}

	serr := SendMsg(m.virtIOOut, rsp)
	if serr != nil {
		Log("replying response to machine %s failed", m.id)
		os.Exit(1)
	}

	Log("initiation with machine %s completed", m.id)
	fin <- m.idx
}

func waitMachines() {
	fin := make(chan int)

	for idx, machine := range exe.machines {
		machine.idx = idx

		virtIOIn, ierr := net.Dial("unix", machine.virtIOInPath)
		if ierr != nil {
			Log("connecting to %s failed: %s", machine.virtIOInPath, ierr)
			os.Exit(1)
		}
		machine.virtIOIn = virtIOIn

		virtIOOut, oerr := net.Dial("unix", machine.virtIOOutPath)
		if oerr != nil {
			Log("connecting to %s failed: %s", machine.virtIOOutPath, oerr)
			os.Exit(1)
		}
		machine.virtIOOut = virtIOOut

		Log("connected to machine %s", machine.id)

		machine.eventReqRecv = make(chan *I2GMsgReq)
		machine.eventRspSend = make(chan *I2GMsgRsp)

		go machineInitiation(machine, fin)
	}

	Log("connected to all machines")

	for nrInitiated := 0; nrInitiated < len(exe.machines); nrInitiated++ {
		initiatedIdx := <-fin
		Log("machine %s completed initiation", exe.machines[initiatedIdx].id)
	}

	Log("all machines initiated")
}

func waitProcessesDirect(exe *execution) {
	processes := exe.processes

	for nrAccepted := 0; nrAccepted < len(processes); nrAccepted++ {
		n := processes[nrAccepted]

		Log("accepting connection")
		conn, aerr := exe.directListen.Accept()
		if aerr != nil {
			Log("failed to accept on %v: %s", exe.directListen, aerr)
			os.Exit(1)
		}
		Log("accepted connection: %v", conn)

		n.conn = conn

		n.idx = nrAccepted
		n.state = PROCESS_STATE_CONNECTED
		n.startExecution = make(chan interface{})
		n.endExecution = make(chan bool)
		n.endCompletion = make(chan interface{})
		n.eventReqToMain = make(chan *I2GMsgReq_Event)
		n.gotoNext = make(chan interface{})
		n.exe = exe

		n.eventReqRecv = make(chan *I2GMsgReq)
		n.eventRspSend = make(chan *I2GMsgRsp)

		n.directRecvEndCompletion = make(chan interface{})
		n.directSendEndCompletion = make(chan interface{})

		go func(n *process) {
			for {
				req := &I2GMsgReq{}

				rerr := RecvMsg(n.conn, req)
				if rerr != nil {
					if rerr == io.EOF {
						Log("received EOF from process idx :%d", n.idx)
						Log("recv routine end (process index :%d)", n.idx)
						n.directRecvEndCompletion <- true
						return
					} else {
						Log("failed to recieve request (process index: %d): %s", n.idx, rerr)
						n.directRecvEndCompletion <- true
						return // TODO: error handling
					}
				}

				Log("received message from process idx :%d", n.idx)
				n.eventReqRecv <- req
			}
		}(n)

		go func(n *process) {
			for {
				rsp := <-n.eventRspSend
				serr := SendMsg(n.conn, rsp)
				if serr != nil {
					Log("failed to send response (process index: %d): %s", n.idx, serr)
					n.directSendEndCompletion <- true
					return // TODO: error handling
				}
				if *rsp.Res == I2GMsgRsp_END {
					Log("send routine end (process index :%d)", n.idx)
					n.directSendEndCompletion <- true
					return
				}
			}
		}(n)

		go handleProcess(n)
	}

	Log("all processes are accepted, waiting initiation...")

	for nrInitiated := 0; nrInitiated < len(processes); nrInitiated++ {
		initiatedIdx := <-exe.initiationCompletion
		processes[initiatedIdx].state = PROCESS_STATE_READY
		Log("initiation of process: %s succeed", processes[initiatedIdx].id)
	}
}

func runMachineProxy(exe *execution) {
	for _, m := range exe.machines {
		Log("launching proxy goroutines for machine %s", m.id)

		go func(m *machine) {
			for {
				req := &I2GMsgReq{}

				rerr := RecvMsg(m.virtIOIn, req)
				if rerr != nil {
					Log("failed to recieve request (machine ID: %s): %s", m.id, rerr)
					return // TODO: error handling
				}
				Log("received message from machine: %s", m.id)

				sent := false
				for _, process := range m.processes {
					if process.id != *req.ProcessId {
						continue
					}

					if process.state == PROCESS_STATE_CONNECTED {
						Log("before initiation (%s), sending to machine proxy", process.id)
						m.eventReqRecv <- req
					} else {
						Log("sending message to %s", process.id)
						process.eventReqRecv <- req
					}

					sent = true
					break
				}

				if !sent {
					Log("invalid destination process: %s from machine: %s", *req.ProcessId, m.id)
					os.Exit(1)
				}
			}
		}(m)

		go func(m *machine) {
			for {
				rsp := <-m.eventRspSend
				// Log("replying to machine %s, result: %d, GA message ID: %d", m.id, *rsp.Res, *rsp.GaMsgId)
				serr := SendMsg(m.virtIOOut, rsp)
				if serr != nil {
					Log("failed to send response (machine ID: %s): %s", m.id, serr)
				}
			}
		}(m)
	}
}

func waitProcessesNoDirect(exe *execution) {
	for i := 0; i < len(exe.machines); i++ {
		m := exe.machines[i]

		for _, n := range m.processes {
			// it must be initialized before returning of waitProcessesNoDirect()
			n.startExecution = make(chan interface{})
		}

		go func(m *machine) {
			for nrInitiated := 0; nrInitiated < len(m.processes); nrInitiated++ {
				req := <-m.eventReqRecv

				found := false
				process := (*process)(nil)
				for idx, n := range exe.processes {
					if n.id != *req.ProcessId {
						continue
					}

					found = true
					process = exe.processes[idx]
				}

				if !found {
					Log("invalid process is joining to execution: %s", *req.ProcessId)
					os.Exit(1)
				}

				switch process.state {
				case PROCESS_STATE_CONNECTED:
					// do initiation here

					result := I2GMsgRsp_ACK
					gaMsgId := *req.GaMsgId
					rsp := &I2GMsgRsp{
						Res:     &result,
						GaMsgId: &gaMsgId,
					}
					Log("replying to initiation, ga message id: %d", *rsp.GaMsgId)

					m.eventRspSend <- rsp

					process.state = PROCESS_STATE_READY
					process.endExecution = make(chan bool)
					process.endCompletion = make(chan interface{})
					process.eventReqToMain = make(chan *I2GMsgReq_Event)
					process.gotoNext = make(chan interface{})
					process.eventReqRecv = make(chan *I2GMsgReq)
					process.eventRspSend = make(chan *I2GMsgRsp)
					process.exe = exe

					go func() {
						for {
							rsp := <-process.eventRspSend
							Log("forwarding message to machine")
							m.eventRspSend <- rsp
						}
					}()

					go handleProcess(process)

					m.processes = append(m.processes, process)
				case PROCESS_STATE_READY:
					Log("guestagent or inspector is buggy")
					os.Exit(1)

				case PROCESS_STATE_DEAD:
					// is this correct?
					Log("guestagent or inspector is buggy")
					os.Exit(1)

				default:
					Log("invalid state of process: %d", process.state)
					os.Exit(1)
				}
			}
		}(m)
	}

}

type Event struct {
	ProcId string

	EventType  string
	EventParam string
}

type SingleTrace struct {
	EventSequence []Event
}

func recordNewTrace(dir string, info *SearchModeInfo, trace SingleTrace) {
	newTraceId := info.NrCollectedTraces
	info.NrCollectedTraces++

	infoFile, err := os.OpenFile(dir+"/"+SearchModeInfoPath, os.O_WRONLY, 0666)
	if err != nil {
		Log("failed to open file: %s", err)
		os.Exit(1)
	}

	var infoBuf bytes.Buffer
	enc := gob.NewEncoder(&infoBuf)
	eerr := enc.Encode(info)
	if eerr != nil {
		Log("encode failed: %s", eerr)
		os.Exit(1)
	}

	_, werr := infoFile.Write(infoBuf.Bytes())
	if werr != nil {
		Log("updating info file failed: %s", werr)
		os.Exit(1)
	}

	var traceBuf bytes.Buffer
	enc = gob.NewEncoder(&traceBuf)
	eerr = enc.Encode(&trace)
	if eerr != nil {
		Log("encoding trace failed: %s", eerr)
		os.Exit(1)
	}

	tracePath := fmt.Sprintf("%s/%08x", dir, newTraceId)
	Log("new trace path: %s", tracePath)
	traceFile, oerr := os.Create(tracePath)
	if oerr != nil {
		Log("fialed to create a file for new trace: %s", oerr)
	}

	_, werr = traceFile.Write(traceBuf.Bytes())
	if werr != nil {
		Log("writing new trace to file failed: %s", werr)
		os.Exit(1)
	}
}

func singleSearch(exe *execution, dir string, info *SearchModeInfo) {
	for _, n := range exe.processes {
		Log("starting execution of process %s", n.id)
		n.startExecution <- true
	}
	Log("all processes started")

	eventSeq := make([]Event, 0)

	nrLivingProcesses := len(exe.processes)
	for nrLivingProcesses != 0 {
		nIdx := <-exe.eventArrive
		n := exe.processes[nIdx]
		eventReq := <-n.eventReqToMain

		if *eventReq.Type == I2GMsgReq_Event_EXIT {
			nrLivingProcesses--
			Log("process %s is exiting, remaining living pocesses: %d", n.id, nrLivingProcesses)
			n.endExecution <- false
			continue
		}

		// TODO: search policy interface
		e := Event{
			n.id,

			"FuncCall",
			*eventReq.FuncCall.Name,
		}
		eventSeq = append(eventSeq, e)

		Log("process %s: %v", n.id, eventReq)
		n.gotoNext <- true
	}

	newTrace := SingleTrace{
		eventSeq,
	}
	recordNewTrace(dir, info, newTrace)

	Log("gathering end completion notification from goroutines")
	nrEnded := int32(0)
	fin := make(chan interface{})

	for _, n := range exe.processes {
		go func(n *process) {
			Log("reading endCompletion from %s (%v)", n.id, n.endCompletion)
			<-n.endCompletion
			Log("endCompletion from %s", n.id)

			atomic.AddInt32(&nrEnded, 1)
			if int(nrEnded) == len(exe.processes) {
				fin <- true
			}
		}(n)
	}

	<-fin
	Log("gathering end completion notification completed")
}

func readSearchModeDir(dir string) *SearchModeInfo {
	_, serr := os.Stat(dir + "/" + SearchModeInfoPath)
	if os.IsNotExist(serr) {
		_, cerr := os.Create(dir + "/" + SearchModeInfoPath)
		if cerr != nil {
			Log("failed to create SearchModeInfo file")
			os.Exit(1)
		}

		Log("using fresh directory")
		return &SearchModeInfo{
			0,
		}
	}

	file, err := os.Open(dir + "/" + SearchModeInfoPath)
	if err != nil {
		Log("failed to open search mode info: %s", err)
		os.Exit(1)
	}

	fi, serr := file.Stat()
	if serr != nil {
		Log("failed to stat: %s", err)
		os.Exit(1)
	}

	buf := make([]byte, fi.Size())
	_, rerr := file.Read(buf)
	if rerr != nil {
		Log("failed to read: %s", rerr)
		os.Exit(1)
	}

	byteBuf := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(byteBuf)
	var ret SearchModeInfo
	derr := dec.Decode(&ret)
	if derr != nil {
		Log("decode error; %s", derr)
		os.Exit(1)
	}

	Log("a number of collected traces: %d", ret.NrCollectedTraces)
	return &ret
}

func searchMode(flags orchestratorFlags, exe *execution, dir string) {
	info := readSearchModeDir(dir)

	Log("start execution loop body")
	if !exe.globalFlags.direct {
		waitProcessesNoDirect(exe)
		runMachineProxy(exe)
	} else {
		waitProcessesDirect(exe)
	}
	singleSearch(exe, dir, info)
	Log("end execution loop body")

}

func simulationMode(flags orchestratorFlags, exe *execution) {
	for {
		Log("start execution loop body")
		if !exe.globalFlags.direct {
			waitProcessesNoDirect(exe)
			runMachineProxy(exe)
		} else {
			waitProcessesDirect(exe)
		}
		runExecution()
		Log("end execution loop body")
	}

	Log("should not reach here!")
}

func initSearchModeDir(dir string) {
	// should be called without daemonize

	var infoBuf bytes.Buffer
	enc := gob.NewEncoder(&infoBuf)

	des, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("failed to read directory %s: %s\n", dir, err)
		return
	}

	if len(des) != 0 {
		fmt.Printf("directory %s is not empty\n", dir)
		return
	}

	info := SearchModeInfo{
		0,
	}

	err = enc.Encode(&info)
	if err != nil {
		fmt.Printf("failed to encode search mode information: %s\n", err)
		return
	}

	file, oerr := os.Create(dir + "/" + SearchModeInfoPath)
	if oerr != nil {
		fmt.Printf("failed to open file for search mode info: %s\n", oerr)
		return
	}

	_, werr := file.Write(infoBuf.Bytes())
	if werr != nil {
		fmt.Printf("failed to write search mode info: %s\n", werr)
		return
	}
}

func launchOrchestrator(flags orchestratorFlags) {
	if flags.InitSearchModeDir {
		if flags.SearchModeDir == "" {
			fmt.Printf("specify a directory path for search mode")
			return
		}

		fmt.Printf("initializing search mode directory: %s", flags.SearchModeDir)
		initSearchModeDir(flags.SearchModeDir)

		return
	}

	if flags.Daemonize && flags.LogFilePath == "" {
		InitLog("/var/log/earthquake-orchestrator.log")
	} else {
		InitLog(flags.LogFilePath)
	}

	if flags.Daemonize {
		context := new(daemon.Context)
		child, _ := context.Reborn()

		if child != nil {
			return
		} else {
			defer context.Release()
		}
	}

	Log("initializing orchestrator")

	exe = parseExecutionFile(flags.ExecutionFilePath, !flags.SearchMode)
	if exe == nil {
		Log("invalid execution file")
		return
	}

	exe.initiationCompletion = make(chan int)
	exe.eventArrive = make(chan int)
	Log("globalFlags.direct: %d", exe.globalFlags.direct)

	if !exe.globalFlags.direct {
		Log("run in non direct mode (with VMs)")
		waitMachines()

		for i, n := range exe.processes {
			n.idx = i
			n.state = PROCESS_STATE_CONNECTED // FIXME: rename
		}

	} else {
		Log("run in direct mode (no VMs)")

		sport := fmt.Sprintf(":%d", flags.ListenTCPPort)
		ln, lerr := net.Listen("tcp", sport)
		if lerr != nil {
			Log("failed to listen on port %d: %s", flags.ListenTCPPort, lerr)
			os.Exit(1)
		}

		exe.directListen = ln
	}

	if flags.SearchMode {
		Log("search mode, directory: %s", flags.SearchModeDir)
		searchMode(flags, exe, flags.SearchModeDir)
	} else {
		Log("simulation mode")
		simulationMode(flags, exe)
	}
}

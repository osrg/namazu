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
	. "./equtils"
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"net"
	"os"
	"syscall"
)

func initVirtIODeviceFile(path string, r bool) *os.File {
	var file *os.File
	var err error

	if r {
		file, err = os.OpenFile(path, os.O_RDONLY, 0)
	} else {
		file, err = os.OpenFile(path, os.O_WRONLY, 0)
	}

	if err != nil {
		Log("failed to open virtio device file: %s", path)
		os.Exit(1)
	}

	return file
}

func watchWithEpoll(file *os.File, notifyChan chan interface{}) {

	epollfd, err := syscall.EpollCreate(1)
	if err != nil {
		Log("error at EpollCreate()")
		os.Exit(1)
	}

	evt := new(syscall.EpollEvent)
	evt.Events = syscall.EPOLLIN
	evt.Fd = int32(file.Fd())
	err = syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_ADD, int(evt.Fd), evt)
	if err != nil {
		Log("EpollCtl failed: %s", err)
		os.Exit(1)
	}

	for {
		evts := make([]syscall.EpollEvent, 1)

		n, waiterr := syscall.EpollWait(epollfd, evts, -1)
		if waiterr != nil || n != 1 {
			Log("EpollWait failed: %s", waiterr)
			os.Exit(1)
		}

		notifyChan <- true
	}
}

type virtIOChan struct {
	virtIOIn  *os.File
	virtIOOut *os.File

	readReady chan interface{}
}

func (ch *virtIOChan) Read(p []byte) (int, error) {
	<-ch.readReady

	return ch.virtIOIn.Read(p)
}

func (ch *virtIOChan) Write(p []byte) (int, error) {
	return ch.virtIOOut.Write(p)
}

type proc struct {
	idx  int
	conn *net.Conn

	reqCh chan *I2GMsgReq
	rspCh chan *I2GMsgRsp
}

func doProxy(arrive chan int, p *proc) {
	for {
		req := &I2GMsgReq{}
		rerr := RecvMsg(*p.conn, req)
		if rerr != nil {
			Log("receiving message from %s failed: %s", p.conn, rerr)
			os.Exit(1)
		}
		Log("proxy goroutine received message from process (idx: %d)", p.idx)

		arrive <- p.idx
		p.reqCh <- req
		Log("proxy goroutine sent request struct to main goroutine")

		rsp := <-p.rspCh
		Log("orchestrator replied to idx %d", p.idx)
		Log("result: %d", *rsp.Res)
		serr := SendMsg(*p.conn, rsp)
		if serr != nil {
			Log("sending message to %s failed", p.conn)
			os.Exit(1)
		}
	}
}

func launchGuestAgent(flags guestAgentFlags) {
	if flags.MachineID == "" {
		fmt.Printf("specify machine ID\n")
		os.Exit(1)
	}

	InitLog(flags.LogFilePath)

	if flags.Daemonize {
		context := new(daemon.Context)
		child, _ := context.Reborn()

		if child != nil {
			return
		} else {
			defer context.Release()
		}
	}

	virtIOIn := initVirtIODeviceFile(flags.VirtIOPathIn, true)
	notifyCh := make(chan interface{})
	go watchWithEpoll(virtIOIn, notifyCh)
	virtIOOut := initVirtIODeviceFile(flags.VirtIOPathOut, false)

	virtIO := &virtIOChan{
		virtIOIn:  virtIOIn,
		virtIOOut: virtIOOut,
		readReady: notifyCh,
	}

	Log("initialized virtio channel")

	initiation := &G2OMsgReq_Initiation{
		Id: proto.String(flags.MachineID),
	}

	op := G2OMsgReq_INITIATION
	req := &G2OMsgReq{
		Op:         &op,
		Initiation: initiation,
	}

	rsp := &G2OMsgRsp{}
	err := ExecReq(virtIO, req, rsp)
	if err != nil {
		Log("initiation request failed: %s", err)
		os.Exit(1)
	}

	if *rsp.Res != G2OMsgRsp_SUCCEED {
		Log("initiation failed")
		os.Exit(1)
	}
	Log("initiation succeed")

	fromOrchestrator := make(chan *I2GMsgRsp)
	go func() {
		for {
			m := &I2GMsgRsp{}
			oerr := RecvMsg(virtIO, m)
			if oerr != nil {
				Log("receiving response from orchestrator failed: %s", oerr)
				os.Exit(1)
			}

			Log("received message from orchestrator")
			fromOrchestrator <- m
		}
	}()

	serializeCh := make(chan *I2GMsgReq)
	go func() { // goroutine for serializing write to virtIO
		for {
			req := <-serializeCh
			Log("serialize channel received request")
			oerr := SendMsg(virtIO, req)

			if oerr != nil {
				Log("sending request to orchestrator failed: %s", oerr)
				os.Exit(1)
			}
		}
	}()

	procs := make([]*proc, 0)

	reqArrive := make(chan int)
	go func() {
		gaMsgId := 0
		waitingMap := make(map[int]int) // message ID -> index of processes

		for {
			select {
			case procIdx := <-reqArrive:
				nextMsgId := int32(gaMsgId)
				gaMsgId++

				Log("recieve message from process index: %d, guest agent message ID: %d", procIdx, gaMsgId)

				Log("recv chan of proc: %v", procs[procIdx].reqCh)
				waitingMap[int(nextMsgId)] = procIdx
				Log("length of procs: %d", len(procs))
				req := <-procs[procIdx].reqCh
				req.GaMsgId = &nextMsgId

				Log("sending request form process: %d (guest agent message ID: %d)", procIdx, gaMsgId)

				serializeCh <- req
			case oRsp := <-fromOrchestrator:
				if *oRsp.Res == I2GMsgRsp_END {
					Log("end message arrived, finishing guest agent")

					for _, p := range procs {
						res := I2GMsgRsp_END
						endrsp := &I2GMsgRsp{
							Res: &res,
						}
						serr := SendMsg(*p.conn, endrsp)
						if serr != nil {
							Log("sending message to %s failed", *p.conn)
							os.Exit(1)
						}
					}

					os.Exit(0)
				} else {
					recvProcIdx := waitingMap[int(*oRsp.GaMsgId)]
					Log("recvProcIdx: %d", recvProcIdx)
					delete(waitingMap, recvProcIdx)
					Log("orchestrator replied to process: %d", recvProcIdx)

					procs[recvProcIdx].rspCh <- oRsp
				}
			}
		}
	}()

	sport := fmt.Sprintf(":%d", flags.ListenTCPPort)
	ln, lerr := net.Listen("tcp", sport)
	if lerr != nil {
		Log("failed to listen on port %d: %s", flags.ListenTCPPort, lerr)
		os.Exit(1)
	}

	for {
		conn, aerr := ln.Accept()
		if aerr != nil {
			Log("failed to accept on %d: %s", flags.ListenTCPPort, aerr)
			os.Exit(1)
		}

		p := &proc{
			idx:   len(procs),
			conn:  &conn,
			reqCh: make(chan *I2GMsgReq),
			rspCh: make(chan *I2GMsgRsp),
		}
		procs = append(procs, p)

		go doProxy(reqArrive, p)
	}
}

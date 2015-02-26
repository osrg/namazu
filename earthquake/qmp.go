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

package main

import (
	. "./equtils"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func initiateQMP(m *machine) {
	Log("initiating QMP with machine %s", m.id)

	sport := fmt.Sprintf(":%d", m.QMPTCPPort)
	QMPConn, derr := net.Dial("tcp", sport)
	if derr != nil {
		Log("failed to connect to TCP port %d (QEMU process machine ID: %s): %s", m.QMPTCPPort, m.id, derr)
		os.Exit(1)
	}
	m.QMPConn = QMPConn

	qmpInitiateMsg := make(map[string]string)
	qmpInitiateMsg["execute"] = "qmp_capabilities"

	bytes, err := json.Marshal(qmpInitiateMsg)
	if err != nil {
		Log("failed to marshal json: %s", err)
		os.Exit(1)
	}

	wbytes, werr := m.QMPConn.Write(bytes)
	if werr != nil || wbytes != len(bytes) {
		Log("failed to write initiation QMP message to %s: %s", m.id, werr)
		os.Exit(1)
	}

	readBuf := make([]byte, 4096) // FIXME: length
	rbytes, rerr := m.QMPConn.Read(readBuf)
	if rerr != nil {
		Log("failed to read response of initiation QMP message from %s: %s", m.id, rerr)
		os.Exit(1)
	}
	readBuf = readBuf[:rbytes]

	Log("response: %s", readBuf)

	var rsp *map[string]interface{}

	uerr := json.Unmarshal(readBuf, &rsp)
	if uerr != nil {
		Log("failed to unmarshal response of initiation QMP message from %s: %s", m.id, uerr)
		os.Exit(1)
	}

	Log("response: %s", readBuf)

	// TODO: save version information from response?
}

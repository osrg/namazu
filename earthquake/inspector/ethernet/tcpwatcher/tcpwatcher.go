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

package tcpwatcher

import (
	"fmt"

	"github.com/google/gopacket/layers"
)

// not thread-safe
type TCPWatcher struct {
	lastTCPMap map[string]layers.TCP
}

func New() *TCPWatcher {
	return &TCPWatcher{
		lastTCPMap: make(map[string]layers.TCP),
	}
}

func (this *TCPWatcher) tcpMapKey(ip *layers.IPv4, tcp *layers.TCP) string {
	return fmt.Sprintf("%s:%d-%s:%d", ip.SrcIP, tcp.SrcPort, ip.DstIP, tcp.DstPort)
}

func (this *TCPWatcher) isTCPRetrans0(ip *layers.IPv4, tcp *layers.TCP) bool {
	k := this.tcpMapKey(ip, tcp)
	lastTCP, ok := this.lastTCPMap[k]
	if !ok {
		return false
	}
	seqMatch := tcp.Seq == lastTCP.Seq
	ackMatch := tcp.Ack == lastTCP.Ack
	// we need any more bits?
	bitsMatch := tcp.FIN == lastTCP.FIN &&
		tcp.SYN == lastTCP.SYN &&
		tcp.RST == lastTCP.RST &&
		tcp.PSH == lastTCP.PSH &&
		tcp.ACK == lastTCP.ACK
	return seqMatch && ackMatch && bitsMatch
}

func (this *TCPWatcher) IsTCPRetrans(ip *layers.IPv4, tcp *layers.TCP) bool {
	if tcp == nil {
		return false
	}
	retrans := this.isTCPRetrans0(ip, tcp)
	if !retrans {
		k := this.tcpMapKey(ip, tcp)
		if tcp.RST {
			delete(this.lastTCPMap, k)
		} else {
			this.lastTCPMap[k] = *tcp
		}
	}
	return retrans
}

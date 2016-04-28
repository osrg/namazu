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
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestTCPWatcher(t *testing.T) {
	w := New()
	ip0 := &layers.IPv4{
		SrcIP: net.IP{1, 2, 3, 4},
		DstIP: net.IP{5, 6, 7, 8},
	}
	tcp0 := &layers.TCP{
		SrcPort: layers.TCPPort(4242),
		DstPort: layers.TCPPort(42),
		Seq:     4242,
		Ack:     4242,
	}
	tcp1 := &layers.TCP{
		SrcPort: layers.TCPPort(4242),
		DstPort: layers.TCPPort(42),
		Seq:     4242,
		Ack:     4243,
	}
	tcp2 := &layers.TCP{
		SrcPort: layers.TCPPort(4242),
		DstPort: layers.TCPPort(42),
		Seq:     4242,
		Ack:     4243,
		RST:     true,
	}

	assert.False(t, w.IsTCPRetrans(ip0, tcp0))
	assert.True(t, w.IsTCPRetrans(ip0, tcp0))
	assert.False(t, w.IsTCPRetrans(ip0, tcp1))
	assert.False(t, w.IsTCPRetrans(ip0, tcp2))
	assert.False(t, w.IsTCPRetrans(ip0, nil))
	assert.False(t, w.IsTCPRetrans(nil, nil))
}

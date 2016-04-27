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

package ethernet

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func makeEntityIDs(eth *layers.Ethernet, ip *layers.IPv4, tcp *layers.TCP) (string, string) {
	srcEntityID := "_namazu_unknown_entity"
	dstEntityID := "_namazu_unknown_entity"
	if ip != nil && tcp != nil {
		srcEntityID = fmt.Sprintf("entity-%s:%d", ip.SrcIP, tcp.SrcPort)
		dstEntityID = fmt.Sprintf("entity-%s:%d", ip.DstIP, tcp.DstPort)
	}
	return srcEntityID, dstEntityID
}

func parseEthernetBytes(b []byte) (eth *layers.Ethernet, ip *layers.IPv4, tcp *layers.TCP) {
	packet := gopacket.NewPacket(b, layers.LayerTypeEthernet, gopacket.Default)
	if layer := packet.Layer(layers.LayerTypeEthernet); layer != nil {
		eth, _ = layer.(*layers.Ethernet)
	}
	if layer := packet.Layer(layers.LayerTypeIPv4); layer != nil {
		ip, _ = layer.(*layers.IPv4)
	}
	if layer := packet.Layer(layers.LayerTypeTCP); layer != nil {
		tcp, _ = layer.(*layers.TCP)
	}
	return
}

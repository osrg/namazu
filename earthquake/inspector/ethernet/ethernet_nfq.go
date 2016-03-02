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

	netfilter "github.com/AkihiroSuda/go-netfilter-queue"
	log "github.com/cihub/seelog"
	"github.com/google/gopacket/layers"
	"github.com/osrg/earthquake/earthquake/inspector/ethernet/tcpwatcher"
	"github.com/osrg/earthquake/earthquake/inspector/transceiver"
	"github.com/osrg/earthquake/earthquake/signal"
)

// TODO: support user-written MapPacketToEventFunc
type NFQInspector struct {
	OrchestratorURL  string
	EntityID         string
	NFQNumber        uint16
	EnableTCPWatcher bool
	trans            transceiver.Transceiver
	tcpWatcher       *tcpwatcher.TCPWatcher
}

func (this *NFQInspector) Serve() error {
	log.Debugf("Initializing Ethernet Inspector %#v", this)
	var err error

	if this.EnableTCPWatcher {
		this.tcpWatcher = tcpwatcher.New()
	}

	this.trans, err = transceiver.NewTransceiver(this.OrchestratorURL, this.EntityID)
	if err != nil {
		return err
	}
	this.trans.Start()

	nfq, err := netfilter.NewNFQueue(this.NFQNumber, 256, netfilter.NF_DEFAULT_PACKET_SIZE)
	if err != nil {
		return err
	}
	defer nfq.Close()
	nfpChan := nfq.GetPackets()
	for {
		nfp := <-nfpChan
		ip, tcp := this.decodeNFPacket(nfp)
		// note: tcpwatcher is not thread-safe
		if this.EnableTCPWatcher && this.tcpWatcher.IsTCPRetrans(ip, tcp) {
			nfp.SetVerdict(netfilter.NF_DROP)
			continue
		}
		go func() {
			// can we use queue so as to improve determinism?
			if err := this.onPacket(nfp, ip, tcp); err != nil {
				log.Error(err)
			}
		}()
	}
	// NOTREACHED
}

func (this *NFQInspector) decodeNFPacket(nfp netfilter.NFPacket) (ip *layers.IPv4, tcp *layers.TCP) {
	if layer := nfp.Packet.Layer(layers.LayerTypeIPv4); layer != nil {
		ip, _ = layer.(*layers.IPv4)
	}
	if layer := nfp.Packet.Layer(layers.LayerTypeTCP); layer != nil {
		tcp, _ = layer.(*layers.TCP)
	}
	return
}

func (this *NFQInspector) onPacket(nfp netfilter.NFPacket,
	ip *layers.IPv4, tcp *layers.TCP) error {
	srcEntityID, dstEntityID := makeEntityIDs(nil, ip, tcp)
	event, err := signal.NewPacketEvent(this.EntityID,
		srcEntityID, dstEntityID, map[string]interface{}{})
	if err != nil {
		return err
	}
	actionCh, err := this.trans.SendEvent(event)
	if err != nil {
		return err
	}
	action := <-actionCh
	switch action.(type) {
	case *signal.EventAcceptanceAction:
		nfp.SetVerdict(netfilter.NF_ACCEPT)
	case *signal.PacketFaultAction:
		nfp.SetVerdict(netfilter.NF_DROP)
	default:
		return fmt.Errorf("unknown action %s", action)
	}
	return nil
}

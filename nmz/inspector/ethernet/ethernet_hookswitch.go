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
	"bytes"
	"encoding/json"
	"fmt"

	log "github.com/cihub/seelog"
	"github.com/google/gopacket/layers"
	"github.com/osrg/namazu/nmz/inspector/ethernet/hookswitch"
	"github.com/osrg/namazu/nmz/inspector/ethernet/tcpwatcher"
	"github.com/osrg/namazu/nmz/inspector/transceiver"
	"github.com/osrg/namazu/nmz/signal"
	zmq "github.com/vaughan0/go-zmq"
)

// TODO: support user-written MapPacketToEventFunc
type HookSwitchInspector struct {
	OrchestratorURL   string
	EntityID          string
	HookSwitchZMQAddr string
	EnableTCPWatcher  bool
	trans             transceiver.Transceiver
	zmqChannels       *zmq.Channels
	tcpWatcher        *tcpwatcher.TCPWatcher
	// only for testing
	stopCh chan struct{}
}

func NewHookSwitchInspector(orchestratorURL, entityID, hookswitchZMQAddr string, enableTCPWatcher bool) (*HookSwitchInspector, error) {
	insp := HookSwitchInspector{
		OrchestratorURL:   orchestratorURL,
		EntityID:          entityID,
		HookSwitchZMQAddr: hookswitchZMQAddr,
		EnableTCPWatcher:  enableTCPWatcher,
		stopCh:            make(chan struct{}),
	}
	return &insp, nil
}

func (this *HookSwitchInspector) Serve() error {
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

	zmqSocket, err := zmq.NewSocket(zmq.Pair)
	if err != nil {
		return err
	}
	zmqSocket.Bind(this.HookSwitchZMQAddr)
	defer zmqSocket.Close()
	this.zmqChannels = zmqSocket.Channels()
	for {
		select {
		case msgBytes := <-this.zmqChannels.In():
			meta, ethBytes, err := this.decodeZMQMessageBytes(msgBytes)
			if err != nil {
				log.Error(err)
				continue
			}
			eth, ip, tcp := parseEthernetBytes(ethBytes)
			// note: tcpwatcher is not thread-safe
			if this.EnableTCPWatcher && this.tcpWatcher.IsTCPRetrans(ip, tcp) {
				meta.Op = hookswitch.Drop
				err = this.sendZMQMessage(*meta, nil)
				if err != nil {
					log.Error(err)
				}
				continue
			}
			go func() {
				if err := this.onHookSwitchMessage(*meta, ethBytes, eth, ip, tcp); err != nil {
					log.Error(err)
				}
			}()
		case err := <-this.zmqChannels.Errors():
			return err
		case <-this.stopCh:
			log.Info("Shutting down..")
			return nil
		}
	}
}

func (this *HookSwitchInspector) Shutdown() {
	this.stopCh <- struct{}{}
}

func (this *HookSwitchInspector) decodeZMQMessageBytes(msgBytes [][]byte) (*hookswitch.HookSwitchMeta, []byte, error) {
	if len(msgBytes) != 2 {
		return nil, nil, fmt.Errorf("strange number of parts: %d, msgBytes=%v", len(msgBytes), msgBytes)
	}
	meta := new(hookswitch.HookSwitchMeta)
	eth := msgBytes[1]
	if err := json.NewDecoder(bytes.NewReader(msgBytes[0])).Decode(meta); err != nil {
		return nil, nil, err
	}
	return meta, eth, nil
}

func (this *HookSwitchInspector) onHookSwitchMessage(meta hookswitch.HookSwitchMeta,
	bytes []byte,
	eth *layers.Ethernet, ip *layers.IPv4, tcp *layers.TCP) error {
	srcEntityID, dstEntityID := makeEntityIDs(eth, ip, tcp)
	event, err := signal.NewPacketEvent(this.EntityID,
		srcEntityID, dstEntityID, map[string]interface{}{
			"bytes": bytes,
		})
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
		meta.Op = hookswitch.Accept
	case *signal.PacketFaultAction:
		meta.Op = hookswitch.Drop
	default:
		return fmt.Errorf("unknown action %s", action)
	}
	// ignore original ethBytes, nil is enough
	if err = this.sendZMQMessage(meta, nil); err != nil {
		return err
	}
	return nil
}

func (this *HookSwitchInspector) sendZMQMessage(meta hookswitch.HookSwitchMeta, ethBytes []byte) error {
	if !(meta.Op == hookswitch.Accept || meta.Op == hookswitch.Drop) {
		return fmt.Errorf("bad opcode %s", meta.Op)
	}
	w := new(bytes.Buffer)
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		return err
	}
	this.zmqChannels.Out() <- [][]byte{w.Bytes(), ethBytes}
	return nil
}

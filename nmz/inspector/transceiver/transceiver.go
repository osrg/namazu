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

package transceiver

import (
	"fmt"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/osrg/namazu/nmz/signal"
)

type Transceiver interface {
	SendEvent(event signal.Event) (chan signal.Action, error)
	Start()
	// TODO: there should be also "Shutdown()" (especially for testing)
}

func NewTransceiver(orchestratorURL string, entityID string) (Transceiver, error) {
	if strings.HasPrefix(orchestratorURL, "local://") {
		if entityID != "" {
			log.Debugf("entityID %s is ignored by the local transceiver", entityID)
		}
		return &SingletonLocalTransceiver, nil
	} else if strings.HasPrefix(orchestratorURL, "http://") {
		return NewRESTTransceiver(orchestratorURL, entityID)
	} else {
		return nil, fmt.Errorf("strange orchestrator url: %s", orchestratorURL)
	}
}

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

package inspectorhandler

import (
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/entity"
	. "github.com/osrg/earthquake/earthquake/inspectorhandler/localinspectorhandler"
	. "github.com/osrg/earthquake/earthquake/inspectorhandler/pbinspectorhandler"
	. "github.com/osrg/earthquake/earthquake/inspectorhandler/restinspectorhandler"
	. "github.com/osrg/earthquake/earthquake/util/config"
)

type InspectorHandler interface {
	StartAccept(readyEntityCh chan *TransitionEntity)
}

var (
	GlobalLocalInspectorHandler = NewLocalInspectorHandler()
)

func StartInspectorHandlers(readyEntityCh chan *TransitionEntity, cfg Config) {
	GlobalLocalInspectorHandler.StartAccept(readyEntityCh)

	if cfg.IsSet("pbPort") {
		pbPort := cfg.GetInt("pbPort")
		if pbPort > 0 {
			go NewPBInspectorHanlder(pbPort).StartAccept(readyEntityCh)
		} else if pbPort < 0 {
			log.Warnf("ignoring negative pbPort: %d", pbPort)
		}
	}

	if cfg.IsSet("restPort") {
		restPort := cfg.GetInt("restPort")
		if restPort > 0 {
			go NewRESTInspectorHanlder(restPort).StartAccept(readyEntityCh)
		} else if restPort < 0 {
			log.Warnf("ignoring restPort: %d", restPort)
		}
	}

}

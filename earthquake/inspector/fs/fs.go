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

package fs

import (
	"fmt"
	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/signal"
	. "github.com/osrg/earthquake/earthquake/util/rest"
	"github.com/osrg/hookfs/hookfs"
	"syscall"
)

// implements hookfs.HookContext
type EQFSHookContext struct {
	Path string
}

// implements hookfs.Hook
type EQFSHook struct {
	OrchestratorURL string
	EntityID        string
	trans           *Transceiver
}

// implements hookfs.HookWithInit
func (this *EQFSHook) Init() error {
	log.Debugf("Initializing FS Inspector %#v", this)
	var err error
	this.trans, err = NewTransceiver(this.OrchestratorURL, this.EntityID)
	if err != nil {
		return err
	}
	this.trans.Start()
	return nil
}

// implements hookfs.HookOnRead
func (this *EQFSHook) PreRead(path string, length int64, offset int64) ([]byte, error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreRead %s", ctx)
	return nil, nil, false, ctx
}

// implements hookfs.HookOnRead
func (this *EQFSHook) PostRead(realRetCode int32, realBuf []byte, ctx hookfs.HookContext) ([]byte, error, bool) {
	log.Debugf("PostRead %s", ctx)
	path := (ctx.(EQFSHookContext)).Path
	onError := func(err error) ([]byte, error, bool) {
		log.Error(err)
		return nil, nil, false
	}
	event, err := NewFilesystemEvent(this.EntityID, Read, path, map[string]interface{}{})
	if err != nil {
		return onError(err)
	}
	log.Debugf("Event %s", event)
	actionChan, err := this.trans.SendEvent(event)
	if err != nil {
		return onError(err)
	}
	action := <-actionChan
	log.Debugf("Action %s", action)
	switch action.(type) {
	case *EventAcceptanceAction:
		return nil, nil, false
	case *FilesystemFaultAction:
		return nil, syscall.EIO, true
	default:
		return onError(fmt.Errorf("unknown action %s", action))
	}
	// NOTREACHED
}

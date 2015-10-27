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

func commonHook(this *EQFSHook, op FilesystemOp, path string, m map[string]interface{}) (error, bool) {
	event, err := NewFilesystemEvent(this.EntityID, op, path, m)
	if err != nil {
		return err, false
	}
	log.Debugf("Event %s", event)
	actionChan, err := this.trans.SendEvent(event)
	if err != nil {
		return err, false
	}
	action := <-actionChan
	log.Debugf("Action %s", action)
	switch action.(type) {
	case *EventAcceptanceAction:
		return nil, false
	case *FilesystemFaultAction:
		// TODO: get alternative errno from the action
		return syscall.EIO, true
	default:
		return fmt.Errorf("unknown action %s", action), false
	}
	// NOTREACHED
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
	err, hooked := commonHook(this, Read, path, map[string]interface{}{})
	if hooked {
		return nil, err, true
	} else {
		if err != nil {
			log.Error(err)
		}
		return nil, nil, false
	}
	// NOTREACHED
}

// implements hookfs.HookOnMkdir
func (this *EQFSHook) PreMkdir(path string, mode uint32) (error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreMkdir %s", ctx)
	return nil, false, ctx
}

// implements hookfs.HookOnMkdir
func (this *EQFSHook) PostMkdir(realRetCode int32, ctx hookfs.HookContext) (error, bool) {
	log.Debugf("PostMkdir %s", ctx)
	path := (ctx.(EQFSHookContext)).Path
	err, hooked := commonHook(this, Mkdir, path, map[string]interface{}{})
	if hooked {
		return err, true
	} else {
		if err != nil {
			log.Error(err)
		}
		return nil, false
	}
	// NOTREACHED
}

// implements hookfs.HookOnRmdir
func (this *EQFSHook) PreRmdir(path string) (error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreMkdir %s", ctx)
	return nil, false, ctx
}

// implements hookfs.HookOnRmdir
func (this *EQFSHook) PostRmdir(realRetCode int32, ctx hookfs.HookContext) (error, bool) {
	log.Debugf("PostRmdir %s", ctx)
	path := (ctx.(EQFSHookContext)).Path
	err, hooked := commonHook(this, Rmdir, path, map[string]interface{}{})
	if hooked {
		return err, true
	} else {
		if err != nil {
			log.Error(err)
		}
		return nil, false
	}
	// NOTREACHED
}

// implements hookfs.HookOnOpenDir
func (this *EQFSHook) PreOpenDir(path string) (error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreOpenDir %s", ctx)
	return nil, false, ctx
}

// implements hookfs.HookOnOpenDir
func (this *EQFSHook) PostOpenDir(realRetCode int32, ctx hookfs.HookContext) (error, bool) {
	log.Debugf("PostOpenDir %s", ctx)
	path := (ctx.(EQFSHookContext)).Path
	err, hooked := commonHook(this, OpenDir, path, map[string]interface{}{})
	if hooked {
		return err, true
	} else {
		if err != nil {
			log.Error(err)
		}
		return nil, false
	}
	// NOTREACHED
}

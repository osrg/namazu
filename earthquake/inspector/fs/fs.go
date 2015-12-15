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
type FilesystemInspector struct {
	OrchestratorURL string
	EntityID        string
	trans           *Transceiver
}

func (this *FilesystemInspector) String() string {
	return "EQFSHook"
}

// implements hookfs.HookWithInit
func (this *FilesystemInspector) Init() error {
	log.Debugf("Initializing FS Inspector %#v", this)
	var err error
	this.trans, err = NewTransceiver(this.OrchestratorURL, this.EntityID)
	if err != nil {
		return err
	}
	this.trans.Start()
	return nil
}

func (this *FilesystemInspector) commonHook(op FilesystemOp, path string, m map[string]interface{}) (error, bool) {
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
func (this *FilesystemInspector) PreRead(path string, length int64, offset int64) ([]byte, error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreRead %s", ctx)
	return nil, nil, false, ctx
}

// implements hookfs.HookOnRead
func (this *FilesystemInspector) PostRead(realRetCode int32, realBuf []byte, ctx hookfs.HookContext) ([]byte, error, bool) {
	log.Debugf("PostRead %s", ctx)
	path := (ctx.(EQFSHookContext)).Path
	err, hooked := this.commonHook(PostRead, path, map[string]interface{}{})
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

// implements hookfs.HookOnOpenDir
func (this *FilesystemInspector) PreOpenDir(path string) (error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreOpenDir %s", ctx)
	return nil, false, ctx
}

// implements hookfs.HookOnOpenDir
func (this *FilesystemInspector) PostOpenDir(realRetCode int32, ctx hookfs.HookContext) (error, bool) {
	log.Debugf("PostOpenDir %s", ctx)
	path := (ctx.(EQFSHookContext)).Path
	err, hooked := this.commonHook(PostOpenDir, path, map[string]interface{}{})
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

// implements hookfs.HookOnMkdir
func (this *FilesystemInspector) PreMkdir(path string, mode uint32) (error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreMkdir %s", ctx)
	err, hooked := this.commonHook(PreMkdir, path, map[string]interface{}{})
	if hooked {
		return err, true, ctx
	} else {
		if err != nil {
			log.Error(err)
		}
		return nil, false, ctx
	}
	// NOTREACHED
}

// implements hookfs.HookOnMkdir
func (this *FilesystemInspector) PostMkdir(realRetCode int32, ctx hookfs.HookContext) (error, bool) {
	log.Debugf("PostMkdir %s", ctx)
	return nil, false
}

// implements hookfs.HookOnRmdir
func (this *FilesystemInspector) PreRmdir(path string) (error, bool, hookfs.HookContext) {
	ctx := EQFSHookContext{Path: path}
	log.Debugf("PreMkdir %s", ctx)
	err, hooked := this.commonHook(PreRmdir, path, map[string]interface{}{})
	if hooked {
		return err, true, ctx
	} else {
		if err != nil {
			log.Error(err)
		}
		return nil, false, ctx
	}
	// NOTREACHED
}

// implements hookfs.HookOnRmdir
func (this *FilesystemInspector) PostRmdir(realRetCode int32, ctx hookfs.HookContext) (error, bool) {
	log.Debugf("PostRmdir %s", ctx)
	return nil, false
}

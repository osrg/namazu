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

package signal

import "github.com/satori/go.uuid"

// implements Event
type FilesystemEvent struct {
	BasicEvent
}

type FilesystemOp string

const (
	// read-only ops use posthooks
	PostRead    = "post-read"
	PostOpenDir = "post-opendir"
	// write ops use prehooks
	PreWrite = "pre-write"
	PreMkdir = "pre-mkdir"
	PreRmdir = "pre-rmdir"
	PreFsync = "pre-fsync"
)

func NewFilesystemEvent(entityID string, op FilesystemOp, path string, m map[string]interface{}) (Event, error) {
	event := &FilesystemEvent{}
	event.InitSignal()
	event.SetID(uuid.NewV4().String())
	event.SetEntityID(entityID)
	event.SetType("event")
	event.SetClass("FilesystemEvent")
	event.SetDeferred(true)
	opt := map[string]interface{}{
		"op":   op,
		"path": path,
	}
	for k, v := range m {
		opt[k] = v
	}
	event.SetOption(opt)
	return event, nil
}

// implements Event
func (this *FilesystemEvent) DefaultFaultAction() (Action, error) {
	return NewFilesystemFaultAction(this)
}

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

package container

import (
	"github.com/osrg/hookfs/hookfs"
	"github.com/osrg/namazu/nmz/inspector/fs"
	ocutil "github.com/osrg/namazu/nmz/util/orchestrator"
)

func ServeFSInspector(orig, mountpoint string) error {
	hook := fs.FilesystemInspector{
		OrchestratorURL: ocutil.LocalOrchestratorURL,
		EntityID:        "_namazu_container_fs_inspector_" + orig,
	}
	fs, err := hookfs.NewHookFs(orig, mountpoint, &hook)
	if err != nil {
		return err
	}
	if err = fs.Serve(); err != nil {
		return err
	}
	// NOTREACHED
	return nil
}

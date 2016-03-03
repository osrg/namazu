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
	"flag"
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/osrg/earthquake/earthquake/endpoint/local"
	"github.com/osrg/earthquake/earthquake/signal"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/osrg/earthquake/earthquake/util/mockorchestrator"
	"github.com/osrg/hookfs/hookfs"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	orcActionCh := make(chan signal.Action)
	orcEventCh := local.SingletonLocalEndpoint.Start(orcActionCh)
	defer local.SingletonLocalEndpoint.Shutdown()
	mockOrc := mockorchestrator.NewMockOrchestrator(orcEventCh, orcActionCh)
	mockOrc.Start()
	defer mockOrc.Shutdown()
	os.Exit(m.Run())
}

func TestFilesystemInspectorInterfaceImpl(t *testing.T) {
	h := &FilesystemInspector{}
	func(x hookfs.HookWithInit) {}(h)
	func(x hookfs.HookOnRead) {}(h)
	func(x hookfs.HookOnMkdir) {}(h)
	func(x hookfs.HookOnRmdir) {}(h)
	func(x hookfs.HookOnOpenDir) {}(h)
}

func newFUSEServer(t *testing.T, fs *hookfs.HookFs) *fuse.Server {
	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}
	pathFs := pathfs.NewPathNodeFs(fs, nil)
	conn := nodefs.NewFileSystemConnector(pathFs.Root(), opts)
	originalAbs, _ := filepath.Abs(fs.Original)
	mOpts := &fuse.MountOptions{
		AllowOther: false,
		Name:       fs.FsName,
		FsName:     originalAbs,
	}
	server, err := fuse.NewServer(conn.RawFS(), fs.Mountpoint, mOpts)
	assert.NoError(t, err)
	server.SetDebug(true)
	return server
}

func TestFilesystemInspector_3(t *testing.T) {
	testFilesystemInspector(t, 3)
}

func testFilesystemInspector(t *testing.T, n int) {
	insp := FilesystemInspector{
		OrchestratorURL: "local://",
		EntityID:        "dummy",
	}

	origDir, err := ioutil.TempDir("", "fs-test-orig")
	assert.NoError(t, err)
	defer os.RemoveAll(origDir)
	mountDir, err := ioutil.TempDir("", "fs-test-mount")
	assert.NoError(t, err)
	defer os.RemoveAll(mountDir)

	fs, err := hookfs.NewHookFs(origDir, mountDir, &insp)
	assert.NoError(t, err)
	fuseServer := newFUSEServer(t, fs)
	go fuseServer.Serve()
	fuseServer.WaitMount()
	defer fuseServer.Unmount()

	for i := 0; i < n; i++ {
		path := filepath.Join(mountDir, fmt.Sprintf("dir-%d", i))
		err = os.Mkdir(path, 0777)
		assert.NoError(t, err)
		t.Logf("mkdir %s", path)
		err = os.RemoveAll(path)
		assert.NoError(t, err)
		t.Logf("rmdir %s", path)
	}
}

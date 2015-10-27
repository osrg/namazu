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
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/osrg/hookfs/hookfs"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	os.Exit(m.Run())
}

func TestInterfaceImpl(t *testing.T) {
	h := &EQFSHook{}
	func(x hookfs.HookWithInit) {}(h)
	func(x hookfs.HookOnRead) {}(h)
	func(x hookfs.HookOnMkdir) {}(h)
	func(x hookfs.HookOnRmdir) {}(h)
	func(x hookfs.HookOnOpenDir) {}(h)
}

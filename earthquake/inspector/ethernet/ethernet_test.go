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
	"flag"
	"fmt"
	"github.com/osrg/earthquake/earthquake/endpoint/local"
	"github.com/osrg/earthquake/earthquake/signal"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/osrg/earthquake/earthquake/util/mockorchestrator"
	"github.com/stretchr/testify/assert"
	zmq "github.com/vaughan0/go-zmq"
	"io/ioutil"
	"os"
	"testing"
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

func tempZMQAddr(t *testing.T) string {
	tmpfile, err := ioutil.TempFile("", "test-hookswitch-inspector")
	assert.NoError(t, err)
	path := tmpfile.Name()
	// we don't need the file itself
	os.Remove(path)
	addr := fmt.Sprintf("ipc://%s", path)
	return addr
}

func hookSwitchRequest(t *testing.T, id int) [][]byte {
	meta := fmt.Sprintf("{\"id\":%d}", id)
	eth := "\xff\xff\xff\xff\xff\xff" +
		"\x00\x00\x00\x00\x00\x00" +
		"\x08\x00"
	frame := eth + "dummypayload"
	return [][]byte{[]byte(meta), []byte(frame)}
}

func TestHookSwitchInspector_10(t *testing.T) {
	testHookSwitchInspector(t, 10)
}

func testHookSwitchInspector(t *testing.T, n int) {
	context, err := zmq.NewContext()
	assert.NoError(t, err)
	defer context.Close()
	socket, err := context.Socket(zmq.Pair)
	assert.NoError(t, err)
	defer socket.Close()
	zmqAddr := tempZMQAddr(t)
	insp, err := NewHookSwitchInspector(
		"local://",
		"_dummy_entity_id",
		zmqAddr,
		true)
	assert.NoError(t, err)
	socket.Connect(zmqAddr)

	go func() {
		insp.Serve()
	}()
	defer insp.Shutdown()

	chans := socket.Channels()
	defer chans.Close()
	for i := 0; i < n; i++ {
		req := hookSwitchRequest(t, i)
		chans.Out() <- req
		t.Logf("Sent %d, %v", i, req)
		select {
		case rsp := <-chans.In():
			t.Logf("Received %d, %v", i, rsp)
		case err := <-chans.Errors():
			t.Fatal(err)
		}
	}
}

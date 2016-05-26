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

package proc

import (
	"flag"
	"github.com/osrg/namazu/nmz/endpoint/local"
	"github.com/osrg/namazu/nmz/signal"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"github.com/osrg/namazu/nmz/util/mockorchestrator"
	"github.com/stretchr/testify/assert"
	"os"
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

func TestProcInspector(t *testing.T) {
	pid := os.Getpid()
	insp, err := NewProcInspector("local://", "dummy", pid, 200*time.Millisecond)
	assert.NoError(t, err)
	go func() {
		insp.Serve(nil)
	}()
	defer insp.Shutdown()
	// dummy test at the moment
	// TODO: check whether actions are really generated
	time.Sleep(1 * time.Second)
}

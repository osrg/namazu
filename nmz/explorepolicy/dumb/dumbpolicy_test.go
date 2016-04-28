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

package dumb

import (
	"flag"
	"github.com/osrg/namazu/nmz/signal"
	tester "github.com/osrg/namazu/nmz/util/explorepolicytester"
	logutil "github.com/osrg/namazu/nmz/util/log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	signal.RegisterKnownSignals()
	os.Exit(m.Run())
}

func TestDumbWithPacketEvent_10_2(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, New(), 10, 2, true)
}

func TestDumbWithPacketEvent_10_10(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, New(), 10, 10, true)
}

func TestDumbShouldNotBlockWithPacketEvent_10_2(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, New(), 10, 2, false)
}

func TestDumbShouldNotBlockWithPacketEvent_10_10(t *testing.T) {
	tester.XTestPolicyWithPacketEvent(t, New(), 10, 10, false)
}

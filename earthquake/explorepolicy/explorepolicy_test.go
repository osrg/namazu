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

package explorepolicy

import (
	"flag"
	"github.com/osrg/earthquake/earthquake/explorepolicy/dumb"
	"github.com/osrg/earthquake/earthquake/explorepolicy/random"
	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	RegisterKnownExplorePolicies()
	os.Exit(m.Run())
}

func TestPoliciesAreRegistered(t *testing.T) {
	d, err := CreatePolicy("dumb")
	if err != nil {
		t.Fatal(err)
	}
	assert.IsType(t, &dumb.Dumb{}, d)
	r, err := CreatePolicy("random")
	if err != nil {
		t.Fatal(err)
	}
	assert.IsType(t, &random.Random{}, r)
	x, err := CreatePolicy("thisshouldnotexist")
	assert.Error(t, err)
	assert.Nil(t, x)
}

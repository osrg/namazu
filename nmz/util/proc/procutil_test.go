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
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestProcUtil(t *testing.T) {
	// TODO: support argument as pid
	pid := 1
	testProcUtil(t, pid)
}

func testProcUtil(t *testing.T, pid int) {
	lwps, err := LWPs(pid)
	assert.NoError(t, err)
	sort.Ints(lwps)
	t.Logf("LWPs(%d) = %#v", pid, lwps)

	children, err := Children(pid)
	assert.NoError(t, err)
	sort.Ints(children)
	t.Logf("Children(%d) = %#v", pid, children)

	descendants, err := Descendants(pid)
	assert.NoError(t, err)
	sort.Ints(descendants)
	t.Logf("Descendants(%d) = %#v", pid, descendants)

	descendantLWPs, err := DescendantLWPs(pid)
	assert.NoError(t, err)
	sort.Ints(descendantLWPs)
	t.Logf("DescendantLWPs(%d) = %#v", pid, descendantLWPs)
}

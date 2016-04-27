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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdFactory(t *testing.T) {
	f := NewCmdFactory()
	// there shouldn't any access to the file, actually
	f.SetWorkingDir("/tmp/dummy1")
	f.SetMaterialsDir("/tmp/dummy2")
	cmd := f.CreateCmd("echo 42")
	assert.Contains(t, cmd.Env, "NMZ_WORKING_DIR=/tmp/dummy1")
}

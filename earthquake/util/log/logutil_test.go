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

package log

import (
	"io/ioutil"
	"os"
	"testing"

	seelog "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func testInitLogWithDebug(t *testing.T, debug bool) {
	tmpFile, err := ioutil.TempFile("", "logutil_test")
	assert.NoError(t, err)
	t.Logf("using %s as the log file", tmpFile.Name())
	defer func() {
		t.Logf("removing %s as the log file", tmpFile.Name())
		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)
	}()
	InitLog(tmpFile.Name(), debug)
	seelog.Infof("hello world 1")
	seelog.Debugf("hello world 2")
	seelog.Flush()
	contentBytes, err := ioutil.ReadAll(tmpFile)
	assert.NoError(t, err)
	content := string(contentBytes[:])
	t.Logf("log file (%s) content: \"%s\"", tmpFile.Name(), content)
	assert.Contains(t, content, "hello world 1")
	if debug {
		assert.Contains(t, content, "hello world 2")
	} else {
		assert.NotContains(t, content, "hello world 2")
	}
}

func TestInitLog(t *testing.T) {
	testInitLogWithDebug(t, false)
}

func TestInitLogWithDebug(t *testing.T) {
	testInitLogWithDebug(t, true)
}

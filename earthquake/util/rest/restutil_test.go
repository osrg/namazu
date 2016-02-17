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

package rest

import (
	"errors"
	"flag"
	"net/http/httptest"
	"os"
	"testing"

	logutil "github.com/osrg/earthquake/earthquake/util/log"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	logutil.InitLog("", true)
	os.Exit(m.Run())
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("hello error")
	WriteError(w, err)
	body := w.Body.String()
	t.Logf("body: \"%s\"", body)
	assert.Contains(t, body, "hello error")
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	m := map[string]interface{}{
		"foo": 42,
		"bar": map[string]interface{}{
			"baz": 43,
		},
	}
	err := WriteJSON(w, m)
	if err != nil {
		t.Fatal(err)
	}
	body := w.Body.String()
	t.Logf("body: \"%s\"", body)
	assert.Contains(t, body, "baz")
}

func TestWriteBadJSON(t *testing.T) {
	w := httptest.NewRecorder()
	m := map[int]interface{}{
		42: "foo",
	}
	err := WriteJSON(w, m)
	// err should be "json: unsupported type: map[int]interface {}"
	t.Logf("error is expected here: %s", err)
	assert.Error(t, err)
}

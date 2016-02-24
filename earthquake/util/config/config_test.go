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

package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	cfg := New()
	t.Logf("default config: %s", cfg)
	assert.Equal(t, -1, cfg.GetInt("restPort"))
	assert.Equal(t, -1, cfg.GetInt("RESTport"))
	assert.Equal(t, -1, cfg.GetInt("pbPort"))
	assert.Equal(t, "random", cfg.GetString("explorepolicy"))
	assert.Equal(t, "random", strings.ToLower(cfg.GetString("ExplorePolicy")))
}

func TestNewConfigFromString(t *testing.T) {
	tomlString := `
# dummy
[foo]
  bar = 42
`
	yamlString := `
# dummy
foo:
  bar: 43
`

	jsonString := `
{"foo": {"bar": 44}}
`
	tomlCfg, err := NewFromString(tomlString, "toml")
	assert.NoError(t, err)
	assert.Equal(t, 42, tomlCfg.GetInt("foo.bar"))

	yamlCfg, err := NewFromString(yamlString, "yaml")
	assert.NoError(t, err)
	assert.Equal(t, 43, yamlCfg.GetInt("foo.bar"))

	jsonCfg, err := NewFromString(jsonString, "json")
	assert.NoError(t, err)
	assert.Equal(t, 44, jsonCfg.GetInt("foo.bar"))
}

func TestNewConfigFromFile(t *testing.T) {
	tomlString := `
# dummy
[foo]
  bar = 42
`
	tmpFile, err := ioutil.TempFile("", "config_test")
	assert.NoError(t, err)
	fileName := tmpFile.Name()
	t.Logf("using file %s", tmpFile.Name())
	defer func() {
		t.Logf("removing file %s", fileName)
		err = os.Remove(fileName)
		assert.NoError(t, err)
	}()
	err = ioutil.WriteFile(fileName, []byte(tomlString), 0644)
	assert.NoError(t, err)
	_, err = NewFromFile(fileName)
	t.Logf("error is expected here (due to lack of suffix): %s", err)
	assert.Error(t, err)
	t.Logf("Renaming to %s", fileName+".toml")
	err = os.Rename(fileName, fileName+".toml")
	assert.NoError(t, err)
	fileName = fileName + ".toml"
	cfg, err := NewFromFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, 42, cfg.GetInt("foo.bar"))
}

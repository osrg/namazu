// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

package equtils

import (
	"errors"
	"github.com/spf13/viper"
)

// TODO: support env vars for overriding (https://github.com/spf13/viper#working-with-environment-variables)
// TODO: validate config with JSON schema
func ParseConfigFile(filePath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetDefault("run", "")
	v.SetDefault("clean", "")
	v.SetDefault("validate", "")
	v.SetDefault("explorePolicy", "dumb")
	v.SetDefault("explorePolicyParam", map[string]interface{}{})
	v.SetDefault("storageType", "naive")
	// viper supports JSON, YAML, and TOML
	v.SetConfigFile(filePath)
	err := v.ReadInConfig()
	if err == nil {
		if v.GetString("run") == "" {
			err = errors.New("required field \"run\" is missing")
		}
	}
	return v, err
}

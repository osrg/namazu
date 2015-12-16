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
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	*viper.Viper
}

func New() Config {
	cfg := Config{viper.New()}

	///// INIT and RUN
	// Used for "init" command.
	// e.g. "init.sh"
	cfg.SetDefault("init", "")

	// Used for "run" command.
	// e.g. "run.sh"
	cfg.SetDefault("run", "")

	// Used for "run" command.
	// e.g. "clean.sh"
	cfg.SetDefault("clean", "")

	// Used for "run" command.
	// e.g. "validate.sh"
	cfg.SetDefault("validate", "")

	// if true, skip clean.sh when validate.sh failed.
	cfg.SetDefault("notCleanIfValidationFail", false)

	///// STORAGE
	// Used for "run" command
	cfg.SetDefault("storageType", "naive")

	///// EXPLORATION POLICY
	// "earthquake-container" also uses these params
	cfg.SetDefault("explorePolicy", "random")
	cfg.SetDefault("explorePolicyParam", map[string]interface{}{})
	return cfg
}

func NewFromString(s string, typ string) (Config, error) {
	cfg := New()
	cfg.SetConfigType(typ)
	err := viper.ReadConfig(strings.NewReader(s))
	return cfg, err
}

func NewFromFile(filePath string) (Config, error) {
	// viper supports JSON, YAML, and TOML
	cfg := New()
	cfg.SetConfigFile(filePath)
	err := cfg.ReadInConfig()
	return cfg, err
}

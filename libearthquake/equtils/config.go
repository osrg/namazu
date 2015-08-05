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
	"encoding/json"
	"errors"
	"os"
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

// TODO: support env vars for overriding (https://github.com/spf13/viper#working-with-environment-variables)
// TODO: validate config with JSON schema
func ParseConfigFile(filePath string) (*Config, error) {
	v := viper.New()
	cfg := &Config{v}
	cfg.SetDefault("init", "") // e.g. "init.sh"
	cfg.SetDefault("run", "") // e.g. "run.sh"
	cfg.SetDefault("clean", "") // e.g. "clean.sh"
	cfg.SetDefault("validate", "") // e.g. "validate.sh"
	cfg.SetDefault("kill", "") // e.g. "kill.sh", unexpected death
	cfg.SetDefault("shutdown", "") // e.g. "shutdown.sh", graceful shutdown
	cfg.SetDefault("explorePolicy", "dumb")
	cfg.SetDefault("explorePolicyParam", map[string]interface{}{})
	cfg.SetDefault("storageType", "naive")
	// Viper Issue: Default value for nested key #71 (https://github.com/spf13/viper/issues/71)
	cfg.SetDefault("inspectorHandler",
		map[string]interface{}{
			"pb": map[string]interface{}{
				// TODO: port
			},
			"rest": map[string]interface{}{
				// TODO: port
			},
		})
	// viper supports JSON, YAML, and TOML
	cfg.SetConfigFile(filePath)
	err := cfg.ReadInConfig()
	if err == nil {
		if cfg.GetString("run") == "" {
			err = errors.New("required field \"run\" is missing")
		}
	}
	return cfg, err
}

func (this *Config) DumpToJsonFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil { return err }
	defer f.Close()
	enc := json.NewEncoder(f)
	var m map[string]interface{}
	err = this.Marshal(&m)
	if err != nil { return err }
	enc.Encode(m)
	return err
}

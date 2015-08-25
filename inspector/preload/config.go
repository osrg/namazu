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

package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Config struct {
	*viper.Viper
}

func NewConfig(jsonString string) (*Config, error) {
	cfg := &Config{viper.New()}
	cfg.SetDefault("max_sleep", time.Duration(0))
	cfg.SetDefault("debug", false)
	cfg.SetConfigType("json")
	err := cfg.ReadConfig(strings.NewReader(jsonString))
	return cfg, err
}

func (this *Config) ToMap() map[string]interface{} {
	var m map[string]interface{}
	err := this.Marshal(&m)
	if err != nil {
		log.WithField("error", err).Error("this should not really happen")
		return nil
	}
	return m
}

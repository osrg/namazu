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

package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

const ConfigEnvVarName = "EQ_LD_PRELOAD"

var (
	config *Config
)

func initLog() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)
}

func initConfig() {
	configString := os.Getenv(ConfigEnvVarName)
	log.WithFields(log.Fields{
		"name":  ConfigEnvVarName,
		"value": configString}).Info("Configuring")
	if configString == "" {
		configString = "{}" // empty json string
	}
	var err error
	config, err = NewConfig(configString)
	if err != nil {
		log.WithField("error", err).Error("Configuration failed")
	} else {
		log.WithField("config", config.ToMap()).Info("Configuration done")
	}
}

func init() {
	initLog()
	log.Info("Initializing Earthquake Preloadable Syscall Inspector")
	initConfig()
	if config.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
}

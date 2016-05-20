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

// Package core provides version information, debug information, and the initializer
package core

import (
	"errors"
	"os"

	log "github.com/cihub/seelog"
	"github.com/osrg/namazu/nmz/explorepolicy"
	"github.com/osrg/namazu/nmz/signal"
	logutil "github.com/osrg/namazu/nmz/util/log"
)

const NamazuVersion = "0.2.1-SNAPSHOT"

// Returns true if NMZ_DEBUG is set
func DebugMode() bool {
	return os.Getenv("NMZ_DEBUG") != ""
}

// Initializes the Namazu system
func Init() {
	debug := DebugMode()
	logutil.InitLog("", debug)
	signal.RegisterKnownSignals()
	explorepolicy.RegisterKnownExplorePolicies()
}

// Prints information on panic
// Also prints the stack trace if it is in the debug mode (NMZ_DEBUG)
func Recoverer() {
	debug := DebugMode()
	if r := recover(); r != nil {
		log.Criticalf("PANIC: %s", r)
		if debug {
			panic(r)
		} else {
			log.Info("Hint: For debug info, please set \"NMZ_DEBUG\" to 1.")
			os.Exit(1)
		}
	}
}

var EthernetInspectorNotBuiltErr = errors.New(
	"Ethernet inspector is disabled in this statically linked binary. " +
		"Please build a dynamically linked binary instead: " +
		"`go get github.com/osrg/namazu/nmz`")

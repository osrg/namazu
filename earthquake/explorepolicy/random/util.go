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

package random

import (
	"fmt"
	"github.com/osrg/earthquake/earthquake/signal"
)

// parses *ProcSetEvent and returns array of PIDs.
//
// due to JSON nature, we use string for PID representation.
func parseProcSetEvent(event *signal.ProcSetEvent) ([]string, error) {
	option := event.Option()
	procs, ok := option["procs"].([]string)
	if !ok {
		// FIXME: this may not work with REST endpoint.
		// we need to convert []interface{} to []string here..?
		return nil, fmt.Errorf("no procs? this should be an implementation error. event=%#v", event)
	}
	return procs, nil
}

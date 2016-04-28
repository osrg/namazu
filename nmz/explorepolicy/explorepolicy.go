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

package explorepolicy

import (
	"fmt"

	log "github.com/cihub/seelog"
)

type PolicyFactory func() ExplorePolicy

var knownPolicies = map[string]PolicyFactory{}

func CreatePolicy(name string) (ExplorePolicy, error) {
	if fn, ok := knownPolicies[name]; ok {
		return fn(), nil
	}
	return nil, fmt.Errorf("unknown policy: %s", name)
}

func RegisterPolicy(name string, fn PolicyFactory) {
	log.Debugf("Registering an exploration policy \"%s\"", name)
	knownPolicies[name] = fn
}

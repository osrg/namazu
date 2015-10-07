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

	. "../equtils"
	"../historystorage"

	"./bfs"
	"./dfs"
	"./dpor"
	"./dumb"
	"./random"

	"./etcd"
	"./zk2172"
)

type ExplorePolicy interface {
	Init(storage historystorage.HistoryStorage, param map[string]interface{})
	Name() string

	GetNextActionChan() chan *Action
	QueueNextEvent(id string, ev *Event)
}

func CreatePolicy(name string) ExplorePolicy {
	switch name {
	case "dumb":
		return dumb.DumbNew()
	case "random":
		return random.RandomNew()
	case "ZK2172":
		return zk2172.ZK2172New()
	case "DFS":
		return dfs.DFSNew()
	case "BFS":
		return bfs.BFSNew()
	case "DPOR":
		return dpor.DPORNew()
	case "etcd":
		return etcd.EtcdNew()
	default:
		fmt.Printf("unknown search policy: %s\n", name)
	}

	return nil
}

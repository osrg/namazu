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
	"github.com/osrg/earthquake/earthquake/historystorage"
	"github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
)

type ExplorePolicy interface {
	// name of the policy
	Name() string

	// should support dynamic reloading
	LoadConfig(cfg config.Config) error

	// policy can read storage, but not expected to perform write ops.
	// orchestrator should write history to the storage.
	SetHistoryStorage(storage historystorage.HistoryStorage) error

	// dequeue action
	GetNextActionChan() chan signal.Action

	// queue event
	QueueEvent(signal.Event)
}

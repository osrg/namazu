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

package historystorage

import (
	"fmt"
	"github.com/osrg/earthquake/earthquake/historystorage/mongodb"
	"github.com/osrg/earthquake/earthquake/historystorage/naive"
	. "github.com/osrg/earthquake/earthquake/signal"
	"github.com/osrg/earthquake/earthquake/util/config"
	. "github.com/osrg/earthquake/earthquake/util/trace"
	"time"
)

const (
	// TODO: we really need to eliminate hard-coded params (config can be yaml)
	StorageTOMLConfigPath string = "config.toml"
)

type HistoryStorage interface {
	CreateStorage()

	Init()
	Close()
	Name() string

	CreateNewWorkingDir() string
	RecordNewTrace(newTrace *SingleTrace)
	RecordResult(successful bool, requiredTime time.Duration) error

	NrStoredHistories() int
	GetStoredHistory(id int) (*SingleTrace, error)

	IsSuccessful(id int) (bool, error)
	GetRequiredTime(id int) (time.Duration, error)

	Search(prefix []Action) []int
	SearchWithConverter(prefix []Action, converter func(actions []Action) []Action) []int
}

func New(name, dirPath string) (HistoryStorage, error) {
	switch name {
	case "naive":
		return naive.New(dirPath), nil
	case "mongodb":
		return mongodb.New(dirPath), nil
	}
	return nil, fmt.Errorf("unknown history storage: %s", name)
}

func LoadStorage(dirPath string) HistoryStorage {
	confPath := dirPath + "/" + StorageTOMLConfigPath

	// TODO: we should not parse config twice. (run should have already parsed the config)
	cfg, err := config.NewFromFile(confPath)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return nil
	}
	storageType := cfg.Get("storageType")

	switch storageType {
	case "naive":
		return naive.New(dirPath)
	default:
		fmt.Printf("unknown history storage: %s\n", storageType)
		return nil
	}
	// NOTREACHED
}

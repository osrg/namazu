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
	"time"

	. "../equtils"
	"./mongodb"
	"./naive"
)

const (
	// TODO: we really need to eliminate hard-coded params (config can be yaml)
	StorageConfigPath string = "config.json"
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

	Search(prefix []Event) []int
	SearchWithConverter(prefix []Event, converter func(events []Event) []Event) []int
}

func New(name, dirPath string) HistoryStorage {
	switch name {
	case "naive":
		return naive.New(dirPath)
	case "mongodb":
		return mongodb.New(dirPath)
	default:
		fmt.Printf("unknown history storage: %s\n", name)
	}

	return nil
}

func LoadStorage(dirPath string) HistoryStorage {
	confPath := dirPath + "/" + StorageConfigPath

	// TODO: we should not parse config twice. (run should have already parsed the config)
	vcfg, err := ParseConfigFile(confPath)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return nil
	}
	storageType := vcfg.Get("storageType")

	switch storageType {
	case "naive":
		return naive.New(dirPath)
	default:
		fmt.Printf("unknown history storage: %s\n", storageType)
		return nil
	}

	return nil
}

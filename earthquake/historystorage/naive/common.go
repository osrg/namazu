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

package naive

// Common stuff of naive history storage

import (
	"time"
)

const (
	searchModeInfoPath = "SearchModeInfo" // relative path of metadata
	resultPath         = "result.json"
)

// type of metadata
type searchModeInfo struct {
	NrCollectedTraces int
}

// type of result, per history
// you have to also update analyzers, if you want to modify this structure
type testResult struct {
	Successful   bool                   `json:"successful"`
	RequiredTime time.Duration          `json:"required_time"`
	Metadata     map[string]interface{} `json:"metadata"` // not yet really supported. intended for tags and so on.
}

// type that implements interface HistoryStorage
type Naive struct {
	dir  string
	info *searchModeInfo

	nextWorkingDir string
}

func (n *Naive) Name() string {
	return "naive"
}

func New(dirPath string) *Naive {
	return &Naive{
		dir: dirPath,
	}
}

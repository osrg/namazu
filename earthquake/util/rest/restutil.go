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

// Package rest provides utilities for REST server
package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/cihub/seelog"
)

// Namazu REST API Root string
const APIRoot = "/api/v3"

// Write err to both of stdout and w
func WriteError(w http.ResponseWriter, err error) {
	log.Errorf("this error will be also written to http: %s", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	w.(http.Flusher).Flush()
}

// Write JSON value to w
func WriteJSON(w http.ResponseWriter, value interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return err
	}
	w.(http.Flusher).Flush()
	return nil
}

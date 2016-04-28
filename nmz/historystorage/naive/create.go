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

import (
	// . "../../equtils"

	"bytes"
	"encoding/gob"
	"fmt"
	"os"
)

// CreateStorage(): used for initializing a directory for naive history storage
// called in a process of "nmz init"
func (n *Naive) CreateStorage() {
	var infoBuf bytes.Buffer
	enc := gob.NewEncoder(&infoBuf)

	info := searchModeInfo{
		0,
	}

	err := enc.Encode(&info)
	if err != nil {
		fmt.Printf("failed to encode search mode information: %s\n", err)
		return
	}

	file, oerr := os.Create(n.dir + "/" + searchModeInfoPath)
	if oerr != nil {
		fmt.Printf("failed to open file for search mode info: %s\n", oerr)
		return
	}

	_, werr := file.Write(infoBuf.Bytes())
	if werr != nil {
		fmt.Printf("failed to write search mode info: %s\n", werr)
		return
	}
}

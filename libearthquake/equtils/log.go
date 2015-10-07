// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

package equtils

// Note: Log library comparison:
//  - standard log: lacks log.Debug()
//  - glog: ditto
//  - logrus: lacks __FILE__, __LINE__

import (
	"fmt"
	log "github.com/cihub/seelog"
	"strings"
)

func InitLog(path string) {

	config := `
<seelog type="sync" minlevel="debug">
    <outputs formatid="main">
        <console/>
        <!-- ${extra} -->
    </outputs>
    <formats>
        <format id="main" format="[EQ-%LEV] %Date(15:04:05.00): %Msg (at %File:%Line) %n"/>
    </formats>
</seelog>`

	if path != "" {
		extra := fmt.Sprintf("<file path=\"%s\"/>", path)
		config = strings.Replace(config, "<!-- ${extra} -->",
			extra, -1)
	}

	logger, err := log.LoggerFromConfigAsBytes([]byte(config))
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)
}

func init() {
	InitLog("")
}

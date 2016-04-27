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

// Package log provides the initializer for cihub/seelog
package log

// Note: Log library comparison:
//  - standard log: lacks log.Debug()
//  - glog: ditto
//  - logrus: lacks __FILE__, __LINE__

import (
	"fmt"
	"strings"

	seelog "github.com/cihub/seelog"
)

var Debug = false

// Initialize cihub/seelog.
// Path can be an empty string.
func InitLog(path string, debug bool) {
	config := `
<seelog type="sync" minlevel="${minlevel}">
    <outputs formatid="main">
        <console/>
        <!-- ${extra} -->
    </outputs>
    <formats>
        <format id="main" format="[NMZ-%LEV] %Date(15:04:05.00): %Msg (at %File:%Line) %n"/>
    </formats>
</seelog>`

	if path != "" {
		extra := fmt.Sprintf("<file path=\"%s\"/>", path)
		config = strings.Replace(config, "<!-- ${extra} -->",
			extra, -1)
	}

	Debug = debug
	if debug {
		config = strings.Replace(config, "${minlevel}", "debug", -1)
	} else {
		config = strings.Replace(config, "${minlevel}", "info", -1)
	}

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(config))
	if err != nil {
		panic(err)
	}
	seelog.ReplaceLogger(logger)
	if path != "" {
		seelog.Debugf("Initialized the logger for %s", path)
	}
}

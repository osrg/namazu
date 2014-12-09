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

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// is log level needed?
type logLevel int

const (
	LOGLEVEL_INFO logLevel = 1
	LOGLEVEL_DBG  logLevel = 2
)

var (
	use_stdout = false
	dst_file   *os.File
)

func Log(format string, v ...interface{}) {
	formatted := fmt.Sprintf(format, v...)

	_, file, line, _ := runtime.Caller(1)
	timestr := time.Now().String()

	if dst_file != nil {
		fmt.Fprintf(dst_file, "%s %s(%d): %s\n", timestr, file, line, formatted)
	} else {
		fmt.Printf("%s %s(%d): %s\n", timestr, file, line, formatted)
	}
}

func InitLog(path string) {
	if path == "" {
		return
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		fmt.Printf("failed to open file %s for logging: %s\n", path, err)
		os.Exit(1)
	}

	dst_file = file
	dst_file.Seek(0, 2)
}

func Panic(format string, v ...interface{}) {
	formatted := fmt.Sprintf("PANIC "+format, v...)

	_, file, line, _ := runtime.Caller(1)
	timestr := time.Now().String()

	if dst_file != nil {
		fmt.Fprintf(dst_file, "%s %s(%d): %s\n", timestr, file, line, formatted)
	} else {
		fmt.Printf("%s %s(%d): %s\n", timestr, file, line, formatted)
	}

	os.Exit(1)
}

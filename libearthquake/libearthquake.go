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

package main

import (
	"C"
	. "./equtils"
)


type libExecution struct {
}

var (
	libExe *libExecution
)

func init() {
	Log("***** libearthquake.so: UNDER REFACTORING, YOU SHOULD NOT USE THIS (please try v0.1 instead) *****")
	Log("libearthquake is planned for exposing some DB-lookup functions to external plugins")
}

//export EQInitCtx
func EQInitCtx(configCString *C.char) int {
	return -1
}


//export EQFreeCtx
func EQFreeCtx() int {
	return 0
}



func main() {
	Log("this dummy main() is required: http://qiita.com/yanolab/items/1e0dd7fd27f19f697285")
}

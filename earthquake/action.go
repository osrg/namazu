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

package main

import (
	. "./equtils"
	"os"
	"os/exec"
)

func actionExecCommand(rawParam interface{}) {
	param := rawParam.(map[string]interface{})
	strCmd := param["command"].(string)
	cmd := exec.Command(strCmd)
	err := cmd.Run()
	if err != nil {
		Log("command %s caused error: %s", strCmd, err)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	Log("command %s executed successfully", strCmd)
}

func doAction(action executionUnitAction) {
	switch action.actionType {
	case "nop":
		Log("nop action")
	case "execCommand":
		Log("execCommand action")
		actionExecCommand(action.param)
	default:
		Panic("unknown action: %s", action.actionType)
	}
}

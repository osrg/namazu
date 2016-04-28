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

package signal

import (
	"fmt"

	log "github.com/cihub/seelog"
	"github.com/osrg/namazu/nmz/util/cmd"
	"github.com/satori/go.uuid"
)

// implements Action, OrchestratorSideAction
type ShellAction struct {
	BasicAction
}

// Execute a shell command (blocking). The action is not expected to be tied with any event.
//
// command: shell command string
//
// comments: arbitrary info for user-written exploration policy (not passed to the command itself).
// e.g., - entity_id(string): entity if that will be killed or shut down
//       - expected(bool): true for expected shutdown, false for unexpected kill
func NewShellAction(command string, comments map[string]interface{}) (Action, error) {
	action := &ShellAction{}
	action.InitSignal()
	action.SetID(uuid.NewV4().String())
	// dummy entity id
	action.SetEntityID("_namazu_shell_action_entity")
	action.SetType("action")
	action.SetClass("ShellAction")
	action.SetOption(map[string]interface{}{
		"command":  command,
		"comments": comments,
	})
	return action, nil
}

// implements OrchestratorSideAction
func (this *ShellAction) OrchestratorSideOnly() bool {
	return true
}

// implements OrchestratorSideAction
func (this *ShellAction) ExecuteOnOrchestrator() error {
	commandStr := this.Option()["command"].(string)
	command := cmd.DefaultFactory.CreateCmd(commandStr)
	if command == nil {
		return fmt.Errorf("got nil while executing %s", this)
	}
	// NOTE: this blocks
	log.Debugf("Starting command %s %s (%#v)", command.Path, command.Args, this.Option)
	err := command.Run()
	log.Debugf("Finished command %s %s (%#v), status: %s", command.Path, command.Args, this.Option, err)
	return err
}

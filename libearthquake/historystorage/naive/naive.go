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
	. "../../equtils"

	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"encoding/json"
	"path"
)

// functions that provides basic functionalities of native history storage
// these are mainly called in a process of "earthquake run"

func (n *Naive) updateSearchModeInfo() {
	infoFile, err := os.OpenFile(n.dir+"/"+searchModeInfoPath, os.O_WRONLY, 0666)
	if err != nil {
		Log("failed to open file: %s", err)
		os.Exit(1)
	}

	var infoBuf bytes.Buffer
	enc := gob.NewEncoder(&infoBuf)
	eerr := enc.Encode(n.info)
	if eerr != nil {
		Log("encode failed: %s", eerr)
		os.Exit(1)
	}

	_, werr := infoFile.Write(infoBuf.Bytes())
	if werr != nil {
		Log("updating info file failed: %s", werr)
		os.Exit(1)
	}
}

func recordAsJSONFile(v interface{}, fileName string) error {
	js, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, js, 0644)
}

func recordJSONAction(i int, act *Action, dir string) {
	if ( act.ActionType != "_JSON") {
		panic(fmt.Errorf("bad action %s", act))
	}

	actJsonName := path.Join(dir, fmt.Sprintf("%d.action.json", i))
	if err := recordAsJSONFile(act.ActionParam, actJsonName); err != nil {
		panic(err)
	}
	evtJsonName := path.Join(dir, fmt.Sprintf("%d.event.json", i))
	if err := recordAsJSONFile(act.Evt.EventParam, evtJsonName); err != nil {
		panic(err)
	}
}

func recordAction(i int, act *Action, dir string) {
	if (act.ActionType == "_JSON") {
		recordJSONAction(i, act, dir)
		return
	}
}

func (n *Naive) RecordNewTrace(newTrace *SingleTrace) {
	var traceBuf bytes.Buffer
	enc := gob.NewEncoder(&traceBuf)
	eerr := enc.Encode(&newTrace)
	if eerr != nil {
		Log("encoding trace failed: %s", eerr)
		os.Exit(1)
	}

	tracePath := fmt.Sprintf("%s/history", n.nextWorkingDir)
	Log("new trace path: %s", tracePath)
	traceFile, oerr := os.Create(tracePath)
	if oerr != nil {
		Log("fialed to create a file for new trace: %s", oerr)
	}

	_, werr := traceFile.Write(traceBuf.Bytes())
	if werr != nil {
		Log("writing new trace to file failed: %s", werr)
		os.Exit(1)
	}

	actionTraceDir := path.Join(n.nextWorkingDir, "actions")
	if err := os.Mkdir(actionTraceDir, 0777); err != nil {
		Log("%s", err)
		os.Exit(1)
	}
	for i, act := range newTrace.ActionSequence {
		recordAction(i, &act, actionTraceDir)
	}
}

func (n *Naive) readSearchModeInfo() *searchModeInfo {
	path := n.dir + "/" + searchModeInfoPath
	file, err := os.Open(path)
	if err != nil {
		Log("failed to open search mode info: %s", err)
		os.Exit(1)
	}

	fi, serr := file.Stat()
	if serr != nil {
		Log("failed to stat: %s", err)
		os.Exit(1)
	}

	buf := make([]byte, fi.Size())
	_, rerr := file.Read(buf)
	if rerr != nil {
		Log("failed to read: %s", rerr)
		os.Exit(1)
	}

	byteBuf := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(byteBuf)
	var ret searchModeInfo
	derr := dec.Decode(&ret)
	if derr != nil {
		Log("decode error; %s", derr)
		os.Exit(1)
	}

	Log("a number of collected traces: %d", ret.NrCollectedTraces)
	return &ret
}

func (n *Naive) CreateNewWorkingDir() string {
	if n.nextWorkingDir != "" {
		fmt.Printf("creating working directory twice\n")
		os.Exit(1)
	}

	newDirPath := fmt.Sprintf("%s/%08x", n.dir, n.info.NrCollectedTraces)

	err := os.Mkdir(newDirPath, 0777)
	if err != nil {
		fmt.Printf("failed to create directory %s: %s\n", newDirPath, err)
		os.Exit(1)
	}

	n.info.NrCollectedTraces++
	n.updateSearchModeInfo()

	n.nextWorkingDir = newDirPath
	return newDirPath
}

func (n *Naive) NrStoredHistories() int {
	return n.info.NrCollectedTraces
}

func (n *Naive) GetStoredHistory(id int) (*SingleTrace, error) {
	path := fmt.Sprintf("%s/%08x/history", n.dir, id)

	encoded, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	byteBuf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(byteBuf)
	var ret SingleTrace
	derr := dec.Decode(&ret)
	if derr != nil {
		return nil, derr
	}

	return &ret, nil
}

func (n *Naive) RecordResult(succeed bool, requiredTime time.Duration) error {
	path := fmt.Sprintf("%s/%s", n.nextWorkingDir, resultPath)

	result := testResult{
		succeed,
		requiredTime,
	}

	var resultBuf bytes.Buffer
	enc := gob.NewEncoder(&resultBuf)
	eerr := enc.Encode(&result)
	if eerr != nil {
		return eerr
	}

	resultFile, oerr := os.Create(path)
	if oerr != nil {
		return oerr
	}

	_, werr := resultFile.Write(resultBuf.Bytes())
	if werr != nil {
		return werr
	}

	return nil
}

func (n *Naive) IsSucceed(id int) (bool, error) {
	path := fmt.Sprintf("%s/%08x/%s", n.dir, id, resultPath)

	encoded, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	byteBuf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(byteBuf)
	var ret testResult
	derr := dec.Decode(&ret)
	if derr != nil {
		return false, derr
	}

	return ret.Succeed, nil

}

func (n *Naive) GetRequiredTime(id int) (time.Duration, error) {
	path := fmt.Sprintf("%s/%08x/%s", n.dir, id, resultPath)

	encoded, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	byteBuf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(byteBuf)
	var ret testResult
	derr := dec.Decode(&ret)
	if derr != nil {
		return 0, derr
	}

	return ret.RequiredTime, nil

}

func (n *Naive) SearchWithConverter(prefix []Event, converter func(events []Event) []Event) []int {
	// FIXME: quite ineffective

	prefixLen := len(prefix)
	matched := make([]int, 0)

	for i := 0; i < n.info.NrCollectedTraces-1; i++ { // FIXME: need to - 1 because the latest trace isn't recorded yet
		history, err := n.GetStoredHistory(i)
		if err != nil {
			Log("failed to get history %i: %s", i, err)
			os.Exit(1)
		}

		if len(history.ActionSequence) < prefixLen {
			continue
		}

		convertee := make([]Event, prefixLen)
		for j, act := range history.ActionSequence[:prefixLen] {
			convertee[j] = *act.Evt
		}

		converted := converter(convertee)
		if AreEventsSliceEqual(prefix, converted) {
			matched = append(matched, i)
		}
	}

	return matched

}

func (n *Naive) Search(prefix []Event) []int {
	// how can we search actions?
	return n.SearchWithConverter(prefix,
		func(events []Event) []Event { return events })
}

func (n *Naive) Init() {
	// required for JSON event and actions
	gob.Register(map[string]interface{}{})
	gob.Register([]map[string]interface{}{})
	n.info = n.readSearchModeInfo()
}

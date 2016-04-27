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
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	log "github.com/cihub/seelog"
	. "github.com/osrg/namazu/nmz/signal"
	signalutil "github.com/osrg/namazu/nmz/util/signal"
	. "github.com/osrg/namazu/nmz/util/trace"
)

// functions that provides basic functionalities of native history storage
// these are mainly called in a process of "earthquake run"

func (n *Naive) updateSearchModeInfo() {
	infoFile, err := os.OpenFile(n.dir+"/"+searchModeInfoPath, os.O_WRONLY, 0666)
	if err != nil {
		panic(log.Criticalf("failed to open file: %s", err))
	}

	var infoBuf bytes.Buffer
	enc := gob.NewEncoder(&infoBuf)
	eerr := enc.Encode(n.info)
	if eerr != nil {
		panic(log.Criticalf("encode failed: %s", eerr))
	}

	_, werr := infoFile.Write(infoBuf.Bytes())
	if werr != nil {
		panic(log.Criticalf("updating info file failed: %s", werr))
	}
}

func recordJSONToFile(v interface{}, fileName string) error {
	// Should we split this from "naive" and move to another storage like "jsonfiles"?
	js, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fileName, js, 0644)
}

func recordAction(i int, act Action, dir string) {
	actJSON := act.JSONMap()
	actJSONName := path.Join(dir, fmt.Sprintf("%d.action.json", i))
	if err := recordJSONToFile(actJSON, actJSONName); err != nil {
		panic(err)
	}

	evt := act.Event()
	if evt != nil {
		evtJSON := evt.JSONMap()
		evtJSONName := path.Join(dir, fmt.Sprintf("%d.event.json", i))
		if err := recordJSONToFile(evtJSON, evtJSONName); err != nil {
			panic(err)
		}
	}
}

func (n *Naive) RecordNewTrace(newTrace *SingleTrace) {
	var traceBuf bytes.Buffer
	enc := gob.NewEncoder(&traceBuf)
	eerr := enc.Encode(&newTrace)
	if eerr != nil {
		panic(log.Criticalf("encoding trace failed: %s", eerr))
	}

	tracePath := fmt.Sprintf("%s/history", n.nextWorkingDir)
	log.Debugf("new trace path: %s", tracePath)
	traceFile, oerr := os.Create(tracePath)
	if oerr != nil {
		panic(log.Criticalf("failed to create a file for new trace: %s", oerr))
	}

	_, werr := traceFile.Write(traceBuf.Bytes())
	if werr != nil {
		panic(log.Criticalf("writing new trace to file failed: %s", werr))
	}

	actionTraceDir := path.Join(n.nextWorkingDir, "actions")
	if err := os.Mkdir(actionTraceDir, 0777); err != nil {
		panic(log.Criticalf("%s", err))
	}
	for i, act := range newTrace.ActionSequence {
		recordAction(i, act, actionTraceDir)
	}
}

func (n *Naive) readSearchModeInfo() *searchModeInfo {
	path := n.dir + "/" + searchModeInfoPath
	file, err := os.Open(path)
	if err != nil {
		panic(log.Criticalf("failed to open search mode info: %s", err))
	}

	fi, serr := file.Stat()
	if serr != nil {
		panic(log.Criticalf("failed to stat: %s", err))
	}

	buf := make([]byte, fi.Size())
	_, rerr := file.Read(buf)
	if rerr != nil {
		panic(log.Criticalf("failed to read: %s", rerr))
	}

	byteBuf := bytes.NewBuffer(buf)
	dec := gob.NewDecoder(byteBuf)
	var ret searchModeInfo
	derr := dec.Decode(&ret)
	if derr != nil {
		panic(log.Criticalf("decode error; %s", derr))
	}

	log.Debugf("a number of collected traces: %d", ret.NrCollectedTraces)
	return &ret
}

func (n *Naive) CreateNewWorkingDir() string {
	if n.nextWorkingDir != "" {
		panic(log.Critical("creating working directory twice"))
	}

	newDirPath := fmt.Sprintf("%s/%08x", n.dir, n.info.NrCollectedTraces)

	err := os.Mkdir(newDirPath, 0777)
	if err != nil {
		panic(log.Criticalf("failed to create directory %s: %s", newDirPath, err))
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

func (n *Naive) RecordResult(successful bool, requiredTime time.Duration) error {
	path := fmt.Sprintf("%s/%s", n.nextWorkingDir, resultPath)

	result := testResult{
		successful,
		requiredTime,
		map[string]interface{}{},
	}
	js, err := json.MarshalIndent(result, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, js, 0644)
}

func (n *Naive) IsSuccessful(id int) (bool, error) {
	path := fmt.Sprintf("%s/%08x/%s", n.dir, id, resultPath)

	encoded, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	byteBuf := bytes.NewBuffer(encoded)

	var ret testResult
	err = json.Unmarshal(byteBuf.Bytes(), &ret)
	if err != nil {
		return false, err
	}
	return ret.Successful, nil
}

func (n *Naive) GetRequiredTime(id int) (time.Duration, error) {
	path := fmt.Sprintf("%s/%08x/%s", n.dir, id, resultPath)

	encoded, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	byteBuf := bytes.NewBuffer(encoded)

	var ret testResult
	err = json.Unmarshal(byteBuf.Bytes(), &ret)
	if err != nil {
		return 0, err
	}

	return ret.RequiredTime, nil
}

func (n *Naive) SearchWithConverter(prefix []Action, converter func(actions []Action) []Action) []int {
	// FIXME: quite ineffective
	matched := make([]int, 0)
	for i := 0; i < n.info.NrCollectedTraces-1; i++ { // FIXME: need to - 1 because the latest trace isn't recorded yet
		history, err := n.GetStoredHistory(i)
		if err != nil {
			panic(log.Criticalf("failed to get history %i: %s", i, err))
		}
		if len(history.ActionSequence) < len(prefix) {
			continue
		}
		converted := converter(history.ActionSequence)
		if signalutil.AreActionsSliceEqual(prefix, converted) {
			matched = append(matched, i)
		}
	}
	return matched
}

func (n *Naive) Search(prefix []Action) []int {
	return n.SearchWithConverter(prefix,
		func(actions []Action) []Action { return actions })
}

func (n *Naive) Init() {
	n.info = n.readSearchModeInfo()
}

func (n *Naive) Close() {
}

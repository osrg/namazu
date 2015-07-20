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
	"encoding/json"	
	"fmt"
	"os"
	 "io/ioutil"
	"path/filepath"
)


type libExecution struct {
	searchDir string
	experimentCount int 
}

var (
	libExe *libExecution
)

func init() {
        Log("***** libearthquake.so: UNDER REFACTORING, YOU SHOULD NOT USE THIS (please try v0.1 instead) *****")
        Log("libearthquake will expose some DB-lookup functions to external plugins")
	Log("init() ENTER")
	libExe = &libExecution{}
	libExe.searchDir = "/tmp/EQ_UNINITIALIZED_SEARCHDIR"
	libExe.experimentCount = 0
	Log("init() LEAVE")
}

//export EQInitCtx
func EQInitCtx(configJsonCString *C.char) int {
	// TODO: refine JSON format for plugins and use equtils.ParseConfigFile
	configJson := C.GoString(configJsonCString)
	jsonBuf := []byte(configJson)
	var root map[string]interface{}
	if err := json.Unmarshal(jsonBuf, &root); err != nil {
		Log("unmarsharing execution file: %s failed (%s)", configJson, err)
		return -1
	}
	globalFlags := root["globalFlags"].(map[string]interface{})
	// direct := int(globalFlags["direct"].(float64))
	// if direct > 0 {
	// 	Log("WARN: direct mode is deprecated, because you can use socat to connect between inside the VM and outside the VM")
	// }
	searchFlags := globalFlags["search"].(map[string]interface{})
	searchDir := searchFlags["directory"].(string)
	Log("searchDir: %s", searchDir)
	if err := os.MkdirAll(searchDir, 0755); err != nil {
		Log("Mkdir error %s", err)
		return -1
	}
	libExe.searchDir = searchDir
	return 0
}


//export EQFreeCtx
func EQFreeCtx() int {
	return 0
}



//export EQRegistExecutionHistory_UnstableAPI
func EQRegistExecutionHistory_UnstableAPI(shortNameCString *C.char, jsonCString *C.char) int {
	nameStr := C.GoString(shortNameCString)
	jsonStr := C.GoString(jsonCString)
	Log("EQRegistExecutionHistory_UnstableAPI: nameStr=%s", nameStr)

	libExe.experimentCount += 1
	dirPath := fmt.Sprintf("%s/history/%016d", libExe.searchDir, libExe.experimentCount)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		Log("Mkdir %s error %s", dirPath, err)
		return -1
	}

	nameFilePath := fmt.Sprintf("%s/name", dirPath)
	Log("EQRegistExecutionHistory_UnstableAPI: writing %s", nameFilePath)	
	if err := ioutil.WriteFile(nameFilePath, []byte(nameStr), 0644); err != nil {
		Log("WriteFile error %s", err)
		return -1
	}
	Log("EQRegistExecutionHistory_UnstableAPI: wrote %s", nameFilePath)		

	jsonFilePath := fmt.Sprintf("%s/json", dirPath)	
	Log("EQRegistExecutionHistory_UnstableAPI: writing %s", jsonFilePath)	
	if err := ioutil.WriteFile(jsonFilePath, []byte(jsonStr), 0644); err != nil {
		Log("WriteFile error %s", err)
		return -1
	}
	Log("EQRegistExecutionHistory_UnstableAPI: wrote %s", jsonFilePath)		
	return 0
}

//export EQGetStatCSV_UnstableAPI
func EQGetStatCSV_UnstableAPI() *C.char {
	csvStr := "#exp_count\tpattern_count\n"

	seenNames := make(map[string]bool)
	nameFiles := fmt.Sprintf("%s/history/*/name", libExe.searchDir)
	files, err := filepath.Glob(nameFiles)
	if err != nil {
		Log("Glob error %s", err)
		return C.CString("#ERROR")
	}

	currentExpCount := 0
	currentPatternCount := 0
	for _, nameFile := range files {
		currentExpCount += 1
		nameBuf, err := ioutil.ReadFile(nameFile)
		if err != nil {
			Log("ReadFile error %s", err)
			return C.CString("#ERROR")			
		}
		nameStr := string(nameBuf)
		Log("loop expCount=%d, nameFile=%s, nameStr=%s, pattenCount=%d", 
			currentExpCount, nameFile, nameStr, currentPatternCount)
		if seenNames[nameStr] {
			Log("nameStr=%s is seen before", nameStr)
		} else {
			Log("nameStr=%s is frontier", nameStr)
			seenNames[nameStr] = true
			currentPatternCount += 1
			csvLine := fmt.Sprintf("%d\t%d\n", currentExpCount, currentPatternCount)
			Log("putting line to csv: %s", csvLine)
			csvStr += csvLine
		}
	}
	cStr :=  C.CString(csvStr)
	return cStr
}



func main() {
	Log("this dummy main() is required: http://qiita.com/yanolab/items/1e0dd7fd27f19f697285")
}

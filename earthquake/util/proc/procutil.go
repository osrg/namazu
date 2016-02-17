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

// Package proc provides utilities for inspecting Linux procfs
package proc

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

// LWPs for pid
func LWPs(pid int) ([]int, error) {
	pids := make([]int, 0)
	files, err := filepath.Glob(fmt.Sprintf("/proc/%d/task/[0-9]*", pid))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		pid, err := strconv.Atoi(filepath.Base(file))
		if err != nil {
			return nil, err
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// children for pid
func Children(pid int) ([]int, error) {
	pids := make([]int, 0)
	content, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/task/%d/children", pid, pid))
	if err != nil {
		return nil, err
	}
	pidstrs := strings.Split(string(content), " ")
	for _, pidstr := range pidstrs {
		if pidstr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidstr)
		if err != nil {
			return nil, err
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// descendants for pid
// can consume stack
func Descendants(pid int) ([]int, error) {
	children, err := Children(pid)
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, nil
	} else {
		result := make([]int, 0)
		for _, child := range children {
			descendants, err := Descendants(child)
			if err != nil {
				return nil, err
			}
			result = extend(result, extend(children, descendants))
		}
		return result, nil
	}
}

// LWPs of descendants + LWPs of itself
// can consume stack
func DescendantLWPs(pid int) ([]int, error) {
	descendants, err := Descendants(pid)
	if err != nil {
		return nil, err
	}
	descendantLWPs, err := LWPs(pid)
	if err != nil {
		return nil, err
	}
	for _, descendant := range descendants {
		lwps, err := LWPs(descendant)
		if err != nil {
			return nil, err
		}
		descendantLWPs = extend(descendantLWPs, lwps)
	}
	return descendantLWPs, nil
}

// unique-extend function as in Python
// http://stackoverflow.com/questions/9251234/go-append-if-unique
func extend(a, b []int) []int {
	set := make(map[int]struct{})
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		set[v] = struct{}{}
	}
	result := make([]int, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}

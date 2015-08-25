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
	// #include <errno.h>
	// static inline int set_errno(int en) { if ( en != 0 ) { errno = en; } }
	"C"
	log "github.com/Sirupsen/logrus"
	"runtime"
	"strings"
	"syscall"
)

func getFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	orig := runtime.FuncForPC(pc).Name() // like "main.open"
	split := strings.Split(orig, ".")
	if len(split) == 2 {
		return split[1] // like "open"
	}
	log.WithField("name", orig).Error("strange function name")
	return orig
}

func getErrno(err error) int32 {
	if err == nil {
		return 0
	} else {
		return int32(err.(syscall.Errno))
	}
}

func exportBase(name string, f func() (int, error), args ...interface{}) int {
	l := log.WithFields(log.Fields{
		"name": name,
		"args": args,
	})
	l.Debug("call")

	prehooked, ret, err := commonPrehook(name, args)
	if !prehooked {
		ret, err = f()
	}
	_, ret, err = commonPosthook(ret, err, name, args)

	errno := getErrno(err)

	l.WithFields(log.Fields{
		"return": ret,
		"errno":  errno,
	}).Debug("return")

	C.set_errno(C.int(errno))
	return ret
}

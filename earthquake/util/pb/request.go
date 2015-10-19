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

package pb

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
)

// http://golang.org/doc/effective_go.html#interfaces

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReaderWriter interface {
	Reader
	Writer
}

// integer header + protobuf marshaled object styled request
// todo: endian
func RecvMsg(rw ReaderWriter, msg proto.Message) error {
	rlengthBuf := make([]byte, 4)
	rbytes := 0

	for rbytes != len(rlengthBuf) {
		r, rerr := rw.Read(rlengthBuf[rbytes:])
		if rerr != nil {
			return rerr
		}

		rbytes += r
	}

	rlength := binary.LittleEndian.Uint32(rlengthBuf)
	recvBuf := make([]byte, int(rlength))

	rbytes = 0
	for rbytes != len(recvBuf) {
		r, rerr := rw.Read(recvBuf[rbytes:])
		if rerr != nil {
			return rerr
		}

		rbytes += r
	}

	uerr := proto.Unmarshal(recvBuf, msg)
	if uerr != nil {
		return uerr
	}

	return nil
}

func SendMsg(rw ReaderWriter, msg proto.Message) error {
	sendBuf, merr := proto.Marshal(msg)
	if merr != nil {
		return merr
	}

	wlength := len(sendBuf)
	wlengthBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(wlengthBuf, uint32(wlength))

	wbytes := 0

	for wbytes != len(wlengthBuf) {
		w, werr := rw.Write(wlengthBuf[wbytes:])
		if werr != nil {
			return werr
		}

		wbytes += w
	}

	wbytes = 0
	for wbytes != len(sendBuf) {
		w, werr := rw.Write(sendBuf[wbytes:])
		if werr != nil {
			return werr
		}

		wbytes += w
	}

	return nil
}

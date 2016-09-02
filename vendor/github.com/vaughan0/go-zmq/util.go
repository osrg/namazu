package zmq

import (
	"fmt"
	"sync"
	"time"
)

// Prefix used for socket-pair endpoint names in MakePair.
var PairPrefix = "socket-pair-"

var counter = 0
var lock sync.Mutex

func nextId() int {
	lock.Lock()
	defer lock.Unlock()
	counter++
	return counter
}

// Creates a pair of connected inproc sockets that can be used for safe inter-thread communication. Returns both
// sockets.
func (c *Context) MakePair() (a *Socket, b *Socket) {
	var err error
	addr := fmt.Sprintf("inproc://%s-%d", PairPrefix, nextId())
	if a, err = c.Socket(Pair); err != nil {
		goto Error
	}
	if err = a.Bind(addr); err != nil {
		goto Error
	}
	if b, err = c.Socket(Pair); err != nil {
		goto Error
	}
	if err = b.Connect(addr); err != nil {
		goto Error
	}
	return

Error:
	if a != nil {
		a.Close()
	}
	if b != nil {
		b.Close()
	}
	panic(err)
}

func toDuration(ival int64, unit time.Duration) time.Duration {
	if ival < 0 {
		return -1
	}
	return time.Duration(ival) * unit
}
func fromDuration(d, unit time.Duration) int64 {
	if d == -1 {
		return -1
	}
	return int64(d / unit)
}

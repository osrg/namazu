package zmq

/*
#include <zmq.h>
*/
import "C"

import (
	"os"
	"time"
)

// An EventSet is a bitmask of IO events.
type EventSet int

const (
	// An input event. Corresponds to receiving on sockets and reading from files.
	In EventSet = C.ZMQ_POLLIN
	// An output event. Corresponds to sending on sockets and writing to files.
	Out = C.ZMQ_POLLOUT
	// An error event. Corresponds to errors on files. The Error event does not apply to sockets.
	Error = C.ZMQ_POLLERR
	// No events
	None = 0
)

// Returns true if the associated socket can receive immediately, or if the associated file can be read from.
func (e EventSet) CanRecv() bool {
	return e&In != 0
}

// Returns true if the associated socket can send immediately, or if the associated file can be written to.
func (e EventSet) CanSend() bool {
	return e&Out != 0
}

// Returns true if the associated file has an error condition. HasError will always return false for sockets.
func (e EventSet) HasError() bool {
	return e&Error != 0
}

// A PollSet represents a set of sockets and/or file descriptors, along with monitored and triggered EventSets. The zero
// PollSet is valid (and empty).
type PollSet struct {
	items []C.zmq_pollitem_t
}

// Adds a Socket to the PollSet along with a set of events to monitor. Returns the index in the PollSet of the added socket.
func (p *PollSet) Socket(sock *Socket, events EventSet) (index int) {
	item := C.zmq_pollitem_t{
		socket: sock.sock,
		events: C.short(events),
	}
	index = len(p.items)
	p.items = append(p.items, item)
	return
}

// Adds a file descriptor to the PollSet along with a set of events to monitor. Returns the index in the PollSet of the
// added file.
func (p *PollSet) Fd(fd uintptr, events EventSet) (index int) {
	item := C.zmq_pollitem_t{
		fd:     C.int(fd),
		events: C.short(events),
	}
	index = len(p.items)
	p.items = append(p.items, item)
	return
}

// Shortcut that calls Fd() with the file's file descriptor.
func (p *PollSet) File(f *os.File, events EventSet) (index int) {
	return p.Fd(f.Fd(), events)
}

// Returns the monitored events for the socket/file of the given index.
func (p *PollSet) Monitored(index int) (events EventSet) {
	return EventSet(p.items[index].events)
}

// Updates the set of monitored events for the socket/file of the given index.
func (p *PollSet) Update(index int, events EventSet) {
	p.items[index].events = C.short(events)
}

// Returns the triggered events for a socket/file of the given index.
func (p *PollSet) Events(index int) EventSet {
	return EventSet(p.items[index].revents)
}

// Poll waits for activity on the monitored set of sockets and/or files. If the timeout is zero, Poll will return
// immediately. If it is negative, Poll will wait forever until an event is triggered. Poll returns the number of
// sockets/files for which events were triggered, or a non-nil error.
func (p *PollSet) Poll(timeout time.Duration) (n int, err error) {
	if p.items == nil || len(p.items) == 0 {
		return
	}
	started := time.Now()
	for {
		var micros C.long
		if timeout < 0 {
			micros = -1
		} else {
			micros = C.long(started.Add(timeout).Sub(time.Now()) / time.Microsecond)
		}
		err = nil
		r := C.zmq_poll(&p.items[0], C.int(len(p.items)), micros)
		if r == -1 {
			err = zmqerr()
		} else {
			n = int(r)
		}
		if err != ErrInterrupted {
			break
		}
	}
	return
}

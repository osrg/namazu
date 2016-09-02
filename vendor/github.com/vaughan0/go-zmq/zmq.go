// Package zmq provides ZeroMQ bindings for Go.
package zmq

/*

#cgo LDFLAGS: -lzmq

#include <zmq.h>
#include <stdlib.h>
#include <string.h>

static int my_errno() {
	return errno;
}

*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

var (
	// ErrTerminated is returned when a socket's context has been closed.
	ErrTerminated = errors.New("zmq context has been terminated")
	// ErrTimeout is returned when an operation times out or a non-blocking operation cannot run immediately.
	ErrTimeout     = errors.New("zmq timeout")
	ErrInterrupted = errors.New("system call interrupted")
)

type SocketType int

const (
	Req    SocketType = C.ZMQ_REQ
	Rep               = C.ZMQ_REP
	Dealer            = C.ZMQ_DEALER
	Router            = C.ZMQ_ROUTER
	Pub               = C.ZMQ_PUB
	Sub               = C.ZMQ_SUB
	XPub              = C.ZMQ_XPUB
	XSub              = C.ZMQ_XSUB
	Push              = C.ZMQ_PUSH
	Pull              = C.ZMQ_PULL
	Pair              = C.ZMQ_PAIR
)

type DeviceType int

const (
	Queue     DeviceType = C.ZMQ_QUEUE
	Forwarder            = C.ZMQ_FORWARDER
	Streamer             = C.ZMQ_STREAMER
)

/* Context */

// A Context manages multiple Sockets. Contexts are thread-safe.
type Context struct {
	ctx unsafe.Pointer
}

// Creates a new Context with the given number of dedicated IO threads.
func NewContextThreads(nthreads int) (ctx *Context, err error) {
	ptr := C.zmq_init(C.int(nthreads))
	if ptr == nil {
		return nil, zmqerr()
	}
	return &Context{ptr}, nil
}

// Creates a new Context with the default number of IO threads (one).
func NewContext() (*Context, error) {
	return NewContextThreads(1)
}

// Closes the Context. Close will block until all related Sockets are closed, and all pending messages are either
// physically transferred to the network or the socket's linger period expires.
func (c *Context) Close() {
	for {
		r := C.zmq_term(c.ctx)
		if r == -1 {
			if C.my_errno() == C.EINTR {
				continue
			}
			panic(zmqerr())
		}
		break
	}
}

// Creates a new Socket of the specified type.
func (c *Context) Socket(socktype SocketType) (sock *Socket, err error) {
	ptr := C.zmq_socket(c.ctx, C.int(socktype))
	if ptr == nil {
		return nil, zmqerr()
	}
	sock = &Socket{
		ctx:  c,
		sock: ptr,
	}
	sock.SetLinger(0)
	return
}

/* Global context */

var (
	globalCtx  *Context = nil
	globalLock sync.Mutex
)

// Returns the default Context. Note that the context will not be created until
// the first call to DefaultContext.
func DefaultContext() *Context {
	globalLock.Lock()
	defer globalLock.Unlock()
	if globalCtx == nil {
		var err error
		if globalCtx, err = NewContext(); err != nil {
			panic(err)
		}
	}
	return globalCtx
}

// Creates a new socket using the default context (see DefaultContext).
func NewSocket(socktype SocketType) (*Socket, error) {
	return DefaultContext().Socket(socktype)
}

/* Socket */

// A ZeroMQ Socket.
type Socket struct {
	ctx  *Context
	sock unsafe.Pointer
}

// Closes the socket.
func (s *Socket) Close() {
	C.zmq_close(s.sock)
}

// Binds the socket to the specified local endpoint address.
func (s *Socket) Bind(endpoint string) (err error) {
	cstr := C.CString(endpoint)
	defer C.free(unsafe.Pointer(cstr))
	r := C.zmq_bind(s.sock, cstr)
	if r == -1 {
		err = zmqerr()
	}
	return
}

// Unbinds the socket from the specified local endpoint address.
func (s *Socket) Unbind(endpoint string) (err error) {
	cstr := C.CString(endpoint)
	defer C.free(unsafe.Pointer(cstr))
	r := C.zmq_unbind(s.sock, cstr)
	if r == -1 {
		err = zmqerr()
	}
	return
}

// Connects the socket to the specified remote endpoint.
func (s *Socket) Connect(endpoint string) (err error) {
	cstr := C.CString(endpoint)
	defer C.free(unsafe.Pointer(cstr))
	r := C.zmq_connect(s.sock, cstr)
	if r == -1 {
		err = zmqerr()
	}
	return
}

// Disconnects the socket from the specified remote endpoint.
func (s *Socket) Disconnect(endpoint string) (err error) {
	cstr := C.CString(endpoint)
	defer C.free(unsafe.Pointer(cstr))
	r := C.zmq_disconnect(s.sock, cstr)
	if r == -1 {
		err = zmqerr()
	}
	return
}

// Sends a single message part. The `more` flag is used to specify whether this is the last part of the message (false),
// or if there are more parts to follow (true). SendPart is fairly low-level, and usually Send will be the preferred
// method to use.
func (s *Socket) SendPart(part []byte, more bool) (err error) {
	for {
		err = nil
		var msg C.zmq_msg_t
		toMsg(&msg, part)
		flags := C.int(0)
		if more {
			flags = C.ZMQ_SNDMORE
		}
		r := C.zmq_msg_send(&msg, s.sock, flags)
		if r == -1 {
			err = zmqerr()
		}
		C.zmq_msg_close(&msg)
		if err != ErrInterrupted {
			break
		}
	}
	return
}

// Sends a message containing a number of parts.
func (s *Socket) Send(parts [][]byte) (err error) {
	for _, part := range parts[:len(parts)-1] {
		if err = s.SendPart(part, true); err != nil {
			return
		}
	}
	return s.SendPart(parts[len(parts)-1], false)
}

// Receives a single part along with a boolean flag (more) indicating whether more parts of the same message follow
// (true), or this is the last part of the message (false). As with Send/SendPart, this is fairly low-level and Recv
// should generally be used instead.
func (s *Socket) RecvPart() (part []byte, more bool, err error) {
	var msg C.zmq_msg_t
	C.zmq_msg_init(&msg)
	for {
		err = nil
		r := C.zmq_msg_recv(&msg, s.sock, 0)
		if r == -1 {
			err = zmqerr()
		}
		if err != ErrInterrupted {
			break
		}
	}
	if err != nil {
		C.zmq_msg_close(&msg)
		return
	}
	part = fromMsg(&msg)
	// Check for more parts
	more = (s.getInt(C.ZMQ_RCVMORE) != 0)
	return
}

// Receives a multi-part message.
func (s *Socket) Recv() (parts [][]byte, err error) {
	parts = make([][]byte, 0)
	for more := true; more; {
		var part []byte
		if part, more, err = s.RecvPart(); err != nil {
			return
		}
		parts = append(parts, part)
	}
	return
}

// Subscribe sets up a filter for incoming messages on Sub sockets.
func (s *Socket) Subscribe(filter []byte) {
	s.setBinary(C.ZMQ_SUBSCRIBE, filter)
}

// Unsubscribes from a filter on a Sub socket.
func (s *Socket) Unsubscribe(filter []byte) {
	s.setBinary(C.ZMQ_UNSUBSCRIBE, filter)
}

/* Device */

// Creates and runs a ZeroMQ Device. See zmq_device(3) for more details.
func Device(deviceType DeviceType, frontend, backend *Socket) {
	C.zmq_device(C.int(deviceType), frontend.sock, backend.sock)
}

// Version reports 0MQ library version.
func Version() (major, minor, patch int) {
	var ma, mi, pa C.int
	C.zmq_version(&ma, &mi, &pa)
	return int(ma), int(mi), int(pa)
}

/* Utilities */

func zmqerr() error {
	eno := C.my_errno()
	switch eno {
	case C.ETERM:
		return ErrTerminated
	case C.EAGAIN:
		return ErrTimeout
	case C.EINTR:
		return ErrInterrupted
	}
	str := C.GoString(C.zmq_strerror(eno))
	return errors.New(str)
}

func toMsg(msg *C.zmq_msg_t, data []byte) {
	C.zmq_msg_init_size(msg, C.size_t(len(data)))
	if len(data) > 0 {
		C.memcpy(C.zmq_msg_data(msg), unsafe.Pointer(&data[0]), C.size_t(len(data)))
	}
}
func fromMsg(msg *C.zmq_msg_t) []byte {
	defer C.zmq_msg_close(msg)
	return C.GoBytes(C.zmq_msg_data(msg), C.int(C.zmq_msg_size(msg)))
}

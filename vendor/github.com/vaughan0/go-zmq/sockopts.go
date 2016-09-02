package zmq

/*
#include <zmq.h>
#include <stdint.h>

#define INT_SIZE (sizeof(int))

*/
import "C"

import (
	"time"
	"unsafe"
)

/* Get */

func (s *Socket) GetType() SocketType {
	return SocketType(s.getInt(C.ZMQ_TYPE))
}
func (s *Socket) GetSendHWM() uint64 {
	return uint64(s.getInt(C.ZMQ_SNDHWM))
}
func (s *Socket) GetRecvHWM() uint64 {
	return uint64(s.getInt(C.ZMQ_RCVHWM))
}
func (s *Socket) GetRecvTimeout() time.Duration {
	return toDuration(int64(s.getInt(C.ZMQ_RCVTIMEO)), time.Millisecond)
}
func (s *Socket) GetSendTimeout() time.Duration {
	return toDuration(int64(s.getInt(C.ZMQ_SNDTIMEO)), time.Millisecond)
}
func (s *Socket) GetAffinity() uint64 {
	return uint64(s.getInt64(C.ZMQ_AFFINITY))
}
func (s *Socket) GetIdentity() []byte {
	return s.getBinary(C.ZMQ_IDENTITY, 256)
}
func (s *Socket) GetRate() (kbits int) {
	return s.getInt(C.ZMQ_RATE)
}
func (s *Socket) GetRecoveryIVL() time.Duration {
	return toDuration(int64(s.getInt(C.ZMQ_RECOVERY_IVL)), time.Millisecond)
}
func (s *Socket) GetSendBuffer() (bytes int) {
	return s.getInt(C.ZMQ_SNDBUF)
}
func (s *Socket) GetRecvBuffer() (bytes int) {
	return s.getInt(C.ZMQ_RCVBUF)
}
func (s *Socket) GetLinger() time.Duration {
	return toDuration(int64(s.getInt(C.ZMQ_LINGER)), time.Millisecond)
}
func (s *Socket) GetReconnectIVL() time.Duration {
	return toDuration(int64(s.getInt(C.ZMQ_RECONNECT_IVL)), time.Millisecond)
}
func (s *Socket) GetReconnectIVLMax() time.Duration {
	return toDuration(int64(s.getInt(C.ZMQ_RECONNECT_IVL_MAX)), time.Millisecond)
}
func (s *Socket) GetBacklog() int {
	return s.getInt(C.ZMQ_BACKLOG)
}
func (s *Socket) GetMaxMsgSize() int64 {
	return s.getInt64(C.ZMQ_MAXMSGSIZE)
}
func (s *Socket) GetMulticastHops() int {
	return s.getInt(C.ZMQ_MULTICAST_HOPS)
}
func (s *Socket) GetIPv4Only() bool {
	return s.getInt(C.ZMQ_IPV4ONLY) != 0
}
func (s *Socket) GetDelayAttachOnConnect() bool {
	return s.getInt(C.ZMQ_DELAY_ATTACH_ON_CONNECT) != 0
}
func (s *Socket) GetFD() uintptr {
	return uintptr(s.getInt(C.ZMQ_FD))
}
func (s *Socket) GetEvents() EventSet {
	return EventSet(s.getInt32(C.ZMQ_EVENTS))
}
func (s *Socket) GetLastEndpoint() string {
	return string(s.getBinary(C.ZMQ_LAST_ENDPOINT, 512))
}
func (s *Socket) GetTCPKeepAlive() int {
	return s.getInt(C.ZMQ_TCP_KEEPALIVE)
}
func (s *Socket) GetTCPKeepAliveIdle() int {
	return s.getInt(C.ZMQ_TCP_KEEPALIVE_IDLE)
}
func (s *Socket) GetTCPKeepAliveCount() int {
	return s.getInt(C.ZMQ_TCP_KEEPALIVE_CNT)
}
func (s *Socket) GetTCPKeepAliveInterval() int {
	return s.getInt(C.ZMQ_TCP_KEEPALIVE_INTVL)
}

/* Set */

func (s *Socket) SetSendHWM(hwm uint64) {
	s.setInt(C.ZMQ_SNDHWM, int(hwm))
}
func (s *Socket) SetRecvHWM(hwm uint64) {
	s.setInt(C.ZMQ_RCVHWM, int(hwm))
}
func (s *Socket) SetAffinity(affinity uint64) {
	s.setInt64(C.ZMQ_AFFINITY, int64(affinity))
}
func (s *Socket) SetIdentitiy(ident []byte) {
	s.setBinary(C.ZMQ_IDENTITY, ident)
}
func (s *Socket) SetRecvTimeout(timeo time.Duration) {
	s.setInt(C.ZMQ_RCVTIMEO, int(fromDuration(timeo, time.Millisecond)))
}
func (s *Socket) SetSendTimeout(timeo time.Duration) {
	s.setInt(C.ZMQ_SNDTIMEO, int(fromDuration(timeo, time.Millisecond)))
}
func (s *Socket) SetRate(kbits int) {
	s.setInt(C.ZMQ_RATE, kbits)
}
func (s *Socket) SetRecoveryIVL(ivl time.Duration) {
	s.setInt(C.ZMQ_RECOVERY_IVL, int(fromDuration(ivl, time.Millisecond)))
}
func (s *Socket) SetSendBuffer(bytes int) {
	s.setInt(C.ZMQ_SNDBUF, bytes)
}
func (s *Socket) SetRecvBuffer(bytes int) {
	s.setInt(C.ZMQ_RCVBUF, bytes)
}
func (s *Socket) SetLinger(linger time.Duration) {
	s.setInt(C.ZMQ_LINGER, int(fromDuration(linger, time.Millisecond)))
}
func (s *Socket) SetReconnectIVL(ivl time.Duration) {
	s.setInt(C.ZMQ_RECONNECT_IVL, int(fromDuration(ivl, time.Millisecond)))
}
func (s *Socket) SetReconnectIVLMax(max time.Duration) {
	s.setInt(C.ZMQ_RECONNECT_IVL_MAX, int(fromDuration(max, time.Millisecond)))
}
func (s *Socket) SetBacklog(backlog int) {
	s.setInt(C.ZMQ_BACKLOG, backlog)
}
func (s *Socket) SetMaxMsgSize(bytes int64) {
	s.setInt64(C.ZMQ_MAXMSGSIZE, bytes)
}
func (s *Socket) SetMulticastHops(ttl int) {
	s.setInt(C.ZMQ_MULTICAST_HOPS, ttl)
}
func (s *Socket) SetIPv4Only(ipv4only bool) {
	s.setBool(C.ZMQ_IPV4ONLY, ipv4only)
}
func (s *Socket) SetDelayAttachOnConnect(delay bool) {
	s.setBool(C.ZMQ_DELAY_ATTACH_ON_CONNECT, delay)
}
func (s *Socket) SetRouterMandatory(errorUnroutable bool) {
	s.setBool(C.ZMQ_ROUTER_MANDATORY, errorUnroutable)
}
func (s *Socket) SetXPubVerbose(verbose bool) {
	s.setBool(C.ZMQ_XPUB_VERBOSE, verbose)
}
func (s *Socket) SetTCPKeepAlive(keepalive int) {
	s.setInt(C.ZMQ_TCP_KEEPALIVE, keepalive)
}
func (s *Socket) SetTCPKeepAliveIdle(idle int) {
	s.setInt(C.ZMQ_TCP_KEEPALIVE_IDLE, idle)
}
func (s *Socket) SetTCPKeepAliveCount(count int) {
	s.setInt(C.ZMQ_TCP_KEEPALIVE_CNT, count)
}
func (s *Socket) SetTCPKeepAliveInterval(interval int) {
	s.setInt(C.ZMQ_TCP_KEEPALIVE_INTVL, interval)
}
func (s *Socket) SetTCPAcceptFilter(filter string) {
	if filter == "" {
		s.setNull(C.ZMQ_TCP_ACCEPT_FILTER)
	} else {
		s.setBinary(C.ZMQ_TCP_ACCEPT_FILTER, []byte(filter))
	}
}

/* Utilities */

func (s *Socket) getInt(opt C.int) int {
	var value C.int
	size := C.size_t(C.INT_SIZE)
	r := C.zmq_getsockopt(s.sock, opt, unsafe.Pointer(&value), &size)
	if r == -1 {
		panic(zmqerr())
	}
	return int(value)
}
func (s *Socket) getInt32(opt C.int) int32 {
	var value C.int32_t
	size := C.size_t(4)
	r := C.zmq_getsockopt(s.sock, opt, unsafe.Pointer(&value), &size)
	if r == -1 {
		panic(zmqerr())
	}
	return int32(value)
}
func (s *Socket) getInt64(opt C.int) int64 {
	var value C.int64_t
	size := C.size_t(8)
	r := C.zmq_getsockopt(s.sock, opt, unsafe.Pointer(&value), &size)
	if r == -1 {
		panic(zmqerr())
	}
	return int64(value)
}
func (s *Socket) getBinary(opt C.int, max int) []byte {
	data := make([]byte, max)
	size := C.size_t(max)
	r := C.zmq_getsockopt(s.sock, opt, unsafe.Pointer(&data[0]), &size)
	if r == -1 {
		panic(zmqerr())
	}
	return data[:int(size)]
}

func (s *Socket) setInt(opt C.int, val int) {
	cval := C.int(val)
	r := C.zmq_setsockopt(s.sock, opt, unsafe.Pointer(&cval), C.INT_SIZE)
	if r == -1 {
		panic(zmqerr())
	}
}
func (s *Socket) setInt64(opt C.int, val int64) {
	r := C.zmq_setsockopt(s.sock, opt, unsafe.Pointer(&val), 8)
	if r == -1 {
		panic(zmqerr())
	}
}
func (s *Socket) setBinary(opt C.int, data []byte) {
	var (
		ptr  unsafe.Pointer
		size C.size_t
	)
	if data != nil && len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
		size = C.size_t(len(data))
	}
	r := C.zmq_setsockopt(s.sock, opt, ptr, size)
	if r == -1 {
		panic(zmqerr())
	}
}
func (s *Socket) setNull(opt C.int) {
	r := C.zmq_setsockopt(s.sock, opt, nil, 0)
	if r == -1 {
		panic(zmqerr())
	}
}
func (s *Socket) setBool(opt C.int, val bool) {
	ival := 0
	if val {
		ival = 1
	}
	s.setInt(opt, ival)
}

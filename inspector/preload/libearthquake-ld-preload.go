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
	unix "golang.org/x/sys/unix"
	"unsafe"
)

// //////////////////// FD Open Operations ////////////////////
//export open
func open(pathnameC *C.char, flagsC C.int, modeC C.int) C.int {
	pathname, flags, mode := C.GoString(pathnameC), int(flagsC), uint32(modeC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				return unix.Open(pathname, flags, mode)
			}, pathname, flags, mode))
}

//export open64
func open64(pathnameC *C.char, flagsC C.int, modeC C.int) C.int {
	return open(pathnameC, flagsC, modeC)
}

//export creat
func creat(pathnameC *C.char, modeC C.int) C.int {
	pathname, mode := C.GoString(pathnameC), uint32(modeC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				return unix.Creat(pathname, mode)
			}, pathname, mode))
}

//export openat
func openat(dirfdC C.int, pathnameC *C.char, flagsC C.int, modeC C.int) C.int {
	dirfd, pathname, flags, mode := int(dirfdC), C.GoString(pathnameC), int(flagsC), uint32(modeC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				return unix.Openat(dirfd, pathname, flags, mode)
			}, dirfd, pathname, flags, mode))
}

//export socket
func socket(domainC C.int, typC C.int, protoC C.int) C.int {
	domain, typ, proto := int(domainC), int(typC), int(protoC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				return unix.Socket(domain, typ, proto)
			}, domain, typ, proto))
}

//export accept
func accept(sockfdC C.int, addr unsafe.Pointer, addrlen unsafe.Pointer) C.int {
	sockfd := int(sockfdC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_ACCEPT, uintptr(sockfd), uintptr(addr), uintptr(addrlen))
				return int(int32(ret)), err
			}, sockfd, addr, addrlen))
}

//export accept4
func accept4(sockfdC C.int, addr unsafe.Pointer, addrlen unsafe.Pointer, flagsC C.int) C.int {
	sockfd, flags := int(sockfdC), int(flagsC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall6(unix.SYS_ACCEPT, uintptr(sockfd), uintptr(addr), uintptr(addrlen), uintptr(flags), 0, 0)
				return int(int32(ret)), err
			}, sockfd, addr, addrlen, flags))
}

//export socketpair
func socketpair(domainC C.int, typC C.int, protoC C.int, sv unsafe.Pointer) C.int {
	domain, typ, proto := int(domainC), int(typC), int(protoC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall6(unix.SYS_SOCKETPAIR, uintptr(domain), uintptr(typ), uintptr(proto), uintptr(sv), 0, 0)
				return int(int32(ret)), err
			}, domain, typ, proto, sv))
}

//export epoll_create
func epoll_create(sizeC C.int) C.int {
	size := int(sizeC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				return unix.EpollCreate(size)
			}, size))
}

//export epoll_create1
func epoll_create1(flagsC C.int) C.int {
	flags := int(flagsC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				return unix.EpollCreate(flags)
			}, flags))
}

//export pipe
func pipe(pipefd unsafe.Pointer) C.int {
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_PIPE, uintptr(pipefd), 0, 0)
				return int(int32(ret)), err
			}, pipefd))
}

//export pipe2
func pipe2(pipefd unsafe.Pointer, flagsC C.int) C.int {
	flags := int(flagsC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_PIPE2, uintptr(pipefd), uintptr(flags), 0)
				return int(int32(ret)), err
			}, pipefd, flags))
}

//export signalfd
func signalfd(fdC C.int, mask unsafe.Pointer, flagsC C.int) C.int {
	fd, flags := int(fdC), int(flagsC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_SIGNALFD, uintptr(fd), uintptr(mask), uintptr(flags))
				return int(int32(ret)), err
			}, fd, mask, flags))
}

//export eventfd
func eventfd(initvalC C.int, flagsC C.int) C.int {
	initval, flags := uint(initvalC), int(flagsC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_EVENTFD, uintptr(initval), uintptr(flags), 0)
				return int(int32(ret)), err
			}, initval, flags))
}

//////////////////// FD Close Operations ////////////////////
func close(fdC C.int) C.int {
	fd := int(fdC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_CLOSE, uintptr(fd), 0, 0)
				return int(int32(ret)), err
			}, fd))
}

//////////////////// EPoll Operations ////////////////////
//export epoll_ctl
func epoll_ctl(epfdC C.int, opC C.int, fdC C.int, event unsafe.Pointer) C.int {
	epfd, op, fd := int(epfdC), int(opC), int(fdC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall6(unix.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(event), 0, 0)
				return int(int32(ret)), err
			}, epfd, op, fd, event))
}

//export epoll_wait
func epoll_wait(epfdC C.int, events unsafe.Pointer, maxeventsC int, timeoutC int) C.int {
	epfd, maxevents, timeout := int(epfdC), int(maxeventsC), int(timeoutC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall6(unix.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(events), uintptr(maxevents), uintptr(timeout), 0, 0)
				return int(int32(ret)), err
			}, epfd, events, maxevents, timeout))
}

//export epoll_pwait
func epoll_pwait(epfdC C.int, events unsafe.Pointer, maxeventsC int, timeoutC int, sigmask unsafe.Pointer) C.int {
	epfd, maxevents, timeout := int(epfdC), int(maxeventsC), int(timeoutC)
	return C.int(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall6(unix.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(events), uintptr(maxevents), uintptr(timeout), uintptr(sigmask), 0)
				return int(int32(ret)), err
			}, epfd, events, maxevents, timeout, sigmask))
}

//////////////////// FD I/O Operations ////////////////////

//export read
func read(fdC C.int, buf unsafe.Pointer, countC C.long) C.long {
	fd, count := int(fdC), int(countC)
	// FIXME: does it work with 32-bit ssize_t?
	return C.long(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_READ, uintptr(fd), uintptr(buf), uintptr(count))
				return int(ret), err
			}, fd, buf, count))
}

//export write
func write(fdC C.int, buf unsafe.Pointer, countC C.long) C.long {
	fd, count := int(fdC), int(countC)
	// FIXME: does it work with 32-bit ssize_t?
	return C.long(
		exportBase(getFuncName(),
			func() (int, error) {
				ret, _, err := unix.Syscall(unix.SYS_WRITE, uintptr(fd), uintptr(buf), uintptr(count))
				return int(ret), err
			}, fd, buf, count))
}

// // TODO: fstat, inotify, pthread, and so on (plus many more..)

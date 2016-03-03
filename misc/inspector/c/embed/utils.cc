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

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <ifaddrs.h>
#include <net/if.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/syscall.h>
#include <sys/eventfd.h>

extern "C" {

#include "utils.h"

int set_nodelay(int fd)
{
  int ret, opt;

  opt = 1;
  ret = setsockopt(fd, IPPROTO_TCP, TCP_NODELAY, &opt, sizeof(opt));
  return ret;
}

int connect_to(const char *name, int port)
{
  char buf[64];
  char hbuf[NI_MAXHOST], sbuf[NI_MAXSERV];
  int fd, ret;
  struct addrinfo hints, *res, *res0;
  struct linger linger_opt = {1, 0};

  memset(&hints, 0, sizeof(hints));
  snprintf(buf, sizeof(buf), "%d", port);

  hints.ai_socktype = SOCK_STREAM;

  ret = getaddrinfo(name, buf, &hints, &res0);
  if (ret) {
    eqi_err("failed to get address info: %m");
    return -1;
  }

  for (res = res0; res; res = res->ai_next) {
    ret = getnameinfo(res->ai_addr, res->ai_addrlen,
		      hbuf, sizeof(hbuf), sbuf, sizeof(sbuf),
		      NI_NUMERICHOST | NI_NUMERICSERV);
    if (ret)
      continue;

    fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
    if (fd < 0)
      continue;

    ret = setsockopt(fd, SOL_SOCKET, SO_LINGER, &linger_opt,
		     sizeof(linger_opt));
    if (ret) {
      eqi_err("failed to set SO_LINGER: %m");
      close(fd);
      continue;
    }

  reconnect:
    ret = connect(fd, res->ai_addr, res->ai_addrlen);
    if (ret) {
      if (errno == EINTR)
	goto reconnect;
      eqi_err("failed to connect to %s:%d: %m", name, port);
      close(fd);
      continue;
    }

    ret = set_nodelay(fd);
    if (ret) {
      eqi_err("%m");
      close(fd);
      break;
    } else
      goto success;
  }
  fd = -1;

 success:
  freeaddrinfo(res0);
  eqi_debug("%d, %s:%d", fd, name, port);

  return fd;
}

int eventfd_xread(int efd)
{
  int ret;
  eventfd_t value = 0;

  do {
    ret = eventfd_read(efd, &value);
  } while (ret < 0 && errno == EINTR);

  if (ret == 0)
    ret = value;
  else if (errno != EAGAIN) {
    eqi_err("eventfd_read() failed, %m");
    exit(1);
  }

  return ret;
}

void eventfd_xwrite(int efd, int value)
{
  int ret;

  do {
    ret = eventfd_write(efd, (eventfd_t)value);
  } while (ret < 0 && (errno == EINTR || errno == EAGAIN));

  if (ret < 0) {
    eqi_err("eventfd_write() failed, %m");
    exit(1);
  }
}

pid_t gettid(void)
{
  return syscall(SYS_gettid);
}

static ssize_t _read(int fd, void *buf, size_t len)
{
  ssize_t nr;
  while (true) {
    nr = read(fd, buf, len);
    if (nr < 0 && (errno == EAGAIN || errno == EINTR))
      continue;
    return nr;
  }
}

static ssize_t _write(int fd, const void *buf, size_t len)
{
  ssize_t nr;
  while (true) {
    nr = write(fd, buf, len);
    if (nr < 0 && (errno == EAGAIN || errno == EINTR))
      continue;
    return nr;
  }
}

ssize_t xread(int fd, void *buf, size_t count)
{
  char *p = (char *)buf;
  ssize_t total = 0;

  while (count > 0) {
    ssize_t loaded = _read(fd, p, count);
    if (loaded < 0)
      return -1;
    if (loaded == 0)
      return total;
    count -= loaded;
    p += loaded;
    total += loaded;
  }

  return total;
}

ssize_t xwrite(int fd, const void *buf, size_t count)
{
  const char *p = (char *)buf;
  ssize_t total = 0;

  while (count > 0) {
    ssize_t written = write(fd, p, count);
    if (written < 0)
      return -1;
    if (!written) {
      errno = ENOSPC;
      return -1;
    }
    count -= written;
    p += written;
    total += written;
  }

  return total;
}

} // extern "C"


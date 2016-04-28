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

#ifndef NMZ_EMBED_UTILS_H
#define NMZ_EMBED_UTILS_H

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
#include <syslog.h>

extern "C" {

#define eqi_err(fmt, args...) \
  syslog(LOG_ERR, "%s(%d): " fmt, __func__, __LINE__, ##args)

#define eqi_debug(fmt, args...) \
  syslog(LOG_DEBUG, "%s(%d): " fmt, __func__, __LINE__, ##args)

#define eqi_info(fmt, args...) \
  syslog(LOG_INFO, "%s(%d): " fmt, __func__, __LINE__, ##args)

int set_nodelay(int fd);
int connect_to(const char *name, int port);

pid_t gettid(void);

void eventfd_xwrite(int efd, int value);
int eventfd_xread(int efd);

ssize_t xread(int fd, void *buf, size_t count);
ssize_t xwrite(int fd, const void *buf, size_t count);

}      /* extern "C" */

#endif	/* NMZ_EMBED_UTILS_H */

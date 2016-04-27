// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

// eq_embed.cpp
// inspected applications must be linked with this file

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

#include <pthread.h>

#include <syslog.h>

#include "inspector_message.pb.h"
#include "utils.h"

#include <vector>

extern "C" {

using namespace std;

struct _atomic_t {
  int val;
};

typedef struct _atomic_t atomic_t;

static inline void atomic_inc(atomic_t *v)
{
  __sync_fetch_and_add(&v->val, 1);
}

#define NMZ_GA_TCP_PORT_ENV_NAME "NMZ_GA_TCP_PORT"
#define NMZ_GA_TCP_PORT_DEFAULT 10000

#define NMZ_DISABLE "NMZ_DISABLE"

#define NMZ_MODE_DIRECT "NMZ_MODE_DIRECT"
#define NMZ_ENV_PROCESS_ID "NMZ_ENV_PROCESS_ID"

static char *_env_processId;

static bool direct_mode;	// if it is true, inspector bypasses guest agent
static int ga_fd;		// fd connected to guest agent
static pthread_t reader_pth;

using namespace equtils;

struct waiting_thread_info {
  pid_t tid;
  int efd;
  int waiting_msg_id;
};

static pthread_mutex_t waiting_thread_mutex = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t waiting_thread_cond = PTHREAD_COND_INITIALIZER;
static vector<struct waiting_thread_info *> waiting_thread_list;

static int recv_msg(InspectorMsgRsp *rsp);

static bool running = true;

static void *reader_thread(void *arg)
{
  int ret;

  eqi_info("reader thread created\n");

  while (true) {
    eqi_info("waiting response from orchestrator...\n");
    InspectorMsgRsp rsp;
    ret = recv_msg(&rsp);
    if (ret) {
      eqi_err("failed to receive InspectorMsgRsp: %m\n");
      exit(1);
    }

    eqi_info("response code from orchestrator: %d\n", rsp.res());

    if (rsp.res() == InspectorMsgRsp_Result_END) {
      eqi_info("inspection ends\n");

      pthread_mutex_lock(&waiting_thread_mutex);
      for (auto ti : waiting_thread_list) {
	eventfd_xwrite(ti->efd, 1);
      }
      pthread_mutex_unlock(&waiting_thread_mutex);

      running = false;
      pthread_exit(0);
    }

    if (rsp.res() != InspectorMsgRsp_Result_ACK) {
      eqi_err("invalid response\n");
      exit(1);
    }

    int msg_id = rsp.msg_id();
    eqi_info("message arrived from orchestrator, message ID: %d\n", msg_id);

    pthread_mutex_lock(&waiting_thread_mutex);
  // retry:
  //   if (waiting_thread_list.size() == 0) {
  //     pthread_cond_wait(&waiting_thread_cond, &waiting_thread_mutex);
  //     goto retry;
  //   }

    int del_pos = 0;

    for (auto ti : waiting_thread_list) {
      if (ti->waiting_msg_id != msg_id) {
	del_pos++;

	continue;
      }

      eqi_info("destination of arrived message: tid == %d\n", ti->tid);
      eventfd_xwrite(ti->efd, 1);

      goto unlock;
    }

    eqi_err("response to unknown message replied, orchestrator is buggy\n");
    exit(1);

  unlock:
    waiting_thread_list.erase(waiting_thread_list.begin() + del_pos);
    pthread_mutex_unlock(&waiting_thread_mutex);
  }
}

static struct waiting_thread_info *get_per_thread_info(void)
{
  static __thread struct waiting_thread_info *info;

  if (info)
    return info;

  info = new struct waiting_thread_info;
  info->efd = eventfd(0, EFD_SEMAPHORE);
  info->tid = gettid();

  return info;
}

static int next_msg_id(void)
{
  static atomic_t id;
  atomic_inc(&id);
  return id.val;
}

static int send_msg(InspectorMsgReq *req);

static void send_event_to_orchestrator(InspectorMsgReq_Event *ev)
{
  pid_t pid, tid;

  pid = getpid();
  tid = gettid();

  InspectorMsgReq req;

  string *env_processId = new string(_env_processId);
  req.set_allocated_process_id(env_processId);
  req.set_type(InspectorMsgReq_Type_EVENT);
  req.set_pid(pid);
  req.set_tid(tid);
  req.set_allocated_event(ev);

  int next_id = next_msg_id();
  req.set_msg_id(next_id);

  struct waiting_thread_info *ti = get_per_thread_info();
  ti->waiting_msg_id = next_id;

  pthread_mutex_lock(&waiting_thread_mutex);
  waiting_thread_list.push_back(ti);
  pthread_mutex_unlock(&waiting_thread_mutex);
  pthread_cond_signal(&waiting_thread_cond);

  int ret = send_msg(&req);
  if (ret) {
    eqi_err("failed to send message to orchestrator: %m\n");
    exit(1);
  }

  eqi_info("reading xevent...\n");
  eventfd_xread(ti->efd);
  eqi_info("read xevent...\n");
  eqi_info("sending event to orchestrator finished\n");
}

void eq_event_func_call(const char *name)
{
  eqi_info("eq_event_func_call() called\n");
  if (!running) {
    eqi_info("inspection already ended\n");
    return;
  }

  InspectorMsgReq_Event *ev = new InspectorMsgReq_Event;
  ev->set_type(InspectorMsgReq_Event_Type_FUNC_CALL);

  InspectorMsgReq_Event_FuncCall *ev_funccall = new InspectorMsgReq_Event_FuncCall;
  ev_funccall->set_name(name);

  ev->set_allocated_funccall(ev_funccall);

  send_event_to_orchestrator(ev);
}

static int send_msg(InspectorMsgReq *req)
{
  string serialized;
  req->SerializeToString(&serialized); // TODO: how to check error of serialization?

  int ret, len = serialized.length();

  ret = xwrite(ga_fd, &len, sizeof(int));
  if (ret < 0) {
    return -1;
  }

  ret = xwrite(ga_fd, serialized.data(), len);
  if (ret < 0) {
    return -1;
  }

  return 0;
}

static int recv_msg(InspectorMsgRsp *rsp)
{
  int ret, len = 0;

  ret = xread(ga_fd, &len, sizeof(int));
  if (ret < 0) {
    return -1;
  }

  char *buf = (char *)calloc(len, sizeof(char));
  if (!buf) {
    return -1;
  }

  ret = xread(ga_fd, buf, len);
  if (ret < 0) {
    return -1;
  }

  string str = buf;
  rsp->ParseFromString(str);

  free(buf);

  return 0;
}

static void exec_req(InspectorMsgReq *req, InspectorMsgRsp *rsp)
{
  int ret;

  ret = send_msg(req);
  if (ret) {
    eqi_err("failed to send message\n");
    exit(1);
  }

  ret = recv_msg(rsp);
  if (ret) {
    eqi_err("failed to receive message\n");
    exit(1);
  }
}

static void initiation(void)
{
  pid_t pid, tid;

  pid = getpid();
  tid = gettid();

  string *env_processId = new string(_env_processId);

  InspectorMsgReq_Initiation *initiation = new InspectorMsgReq_Initiation;
  initiation->set_allocated_processid(env_processId);

  InspectorMsgReq req;

  string *env_processId2 = new string(_env_processId); // FIXME: oops...
  req.set_allocated_process_id(env_processId2);
  req.set_pid(pid);
  req.set_tid(tid);
  req.set_type(InspectorMsgReq_Type_INITIATION);
  req.set_msg_id(0);
  req.set_allocated_initiation(initiation);

  InspectorMsgRsp rsp;
  exec_req(&req, &rsp);

  if (rsp.res() != InspectorMsgRsp_Result_ACK) {
    eqi_err("initiation failed\n");
    exit(1);
  }

  eqi_info("initiation succeed\n");
}

__attribute__((constructor)) void init_namazu_inspection(void)
{
  int tcp_port = NMZ_GA_TCP_PORT_DEFAULT, ret;

  openlog("namazu inspection", LOG_NDELAY | LOG_PID, 0);

  if (getenv(NMZ_DISABLE)) {
    eqi_info("namazu inspection is disabled, do nothing\n");
    running = false;
    return;
  }

  _env_processId = getenv(NMZ_ENV_PROCESS_ID);
  if (!_env_processId) {
    eqi_err("Process ID is required, set environmental variable %s\n", NMZ_ENV_PROCESS_ID);
    exit(1);
  }

  char *env_tcp_port = getenv(NMZ_GA_TCP_PORT_ENV_NAME);
  if (env_tcp_port) {
    tcp_port = atoi(env_tcp_port);
    eqi_debug("specified TCP port of guest agent: %d\n", tcp_port);
  }

  char *env_mode_direct = getenv(NMZ_MODE_DIRECT);
  if (env_mode_direct) {
    direct_mode = true;
    eqi_debug("direct mode\n");
    // in this case, ga_fd connects to orchestrator directly
  }

  ga_fd = connect_to("localhost", tcp_port);
  if (ga_fd < 0) {
    eqi_err("creating socket failed: %m\n");
    exit(1);
  }

  initiation();
  eqi_info("after initiation\n");

  if (pthread_create(&reader_pth, NULL, reader_thread, NULL)) {
    eqi_err("creating reader thread failed: %m\n");
    exit(1);
  }

  eqi_info("constructor ends\n");
}

__attribute__((destructor)) void exit_namazu_inspection(void)
{
  eqi_info("destructor called, process %s is exiting\n", _env_processId);

  InspectorMsgReq_Event *ev = new InspectorMsgReq_Event;
  ev->set_type(InspectorMsgReq_Event_Type_EXIT);
  send_event_to_orchestrator(ev);
}

} // extern "C"


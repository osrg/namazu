#ifndef _LINUXSCHED_H
#define _LINUXSCHED_H

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <linux/unistd.h>
#include <linux/kernel.h>
#include <linux/types.h>
#include <sys/syscall.h>
#include <pthread.h>

#if defined __x86_64__
#define __NR_sched_setattr 314
#define __NR_sched_getattr 315
#elif defined __i386__
#warning not tested on i386
#define __NR_sched_setattr 351
#define __NR_sched_getattr 352
#elif defined __arm__
#warning not tested on arm
#define __NR_sched_setattr 380
#define __NR_sched_getattr 381
#else
#error unknown arch
#endif

struct sched_attr {
  __u32 size;

  __u32 sched_policy;
  __u64 sched_flags;

  /* SCHED_NORMAL, SCHED_BATCH */
  __s32 sched_nice;

  /* SCHED_FIFO, SCHED_RR */
  __u32 sched_priority_;

  /* SCHED_DEADLINE (nsec) */
  __u64 sched_runtime;
  __u64 sched_deadline;
  __u64 sched_period;
};

#define SCHED_ATTR_SIZE (sizeof(struct sched_attr))

/* not available in glibc */
int sched_setattr(pid_t pid,
		  const struct sched_attr *attr,
		  unsigned int flags)
{
  return syscall(__NR_sched_setattr, pid, attr, flags);
}

/* not available in glibc */
int sched_getattr(pid_t pid,
		  struct sched_attr *attr,
		  unsigned int size,
		  unsigned int flags)
{
  return syscall(__NR_sched_getattr, pid, attr, size, flags);
}

#endif /* _LINUXSCHED_H */

package linuxsched

/*
#include "linuxsched.h"
*/
import "C"
import "time"

type SchedPolicy uint32
type SchedFlag uint64

const (
	Normal   SchedPolicy = 0
	FIFO     SchedPolicy = 1
	RR       SchedPolicy = 2
	Batch    SchedPolicy = 3
	Idle     SchedPolicy = 5
	Deadline SchedPolicy = 6

	ResetOnFork SchedFlag = 0x01
)

type SchedAttr struct {
	Policy   SchedPolicy
	Flags    SchedFlag
	Nice     int32
	Priority uint32
	Runtime  time.Duration
	Deadline time.Duration
	Period   time.Duration
}

// sched_setattr(2)
func SetAttr(pid int, attr SchedAttr) error {
	cAttr := C.struct_sched_attr{
		C.__u32(C.SCHED_ATTR_SIZE),
		C.__u32(attr.Policy),
		C.__u64(attr.Flags),
		C.__s32(attr.Nice),
		C.__u32(attr.Priority),
		C.__u64(attr.Runtime.Nanoseconds()),
		C.__u64(attr.Deadline.Nanoseconds()),
		C.__u64(attr.Period.Nanoseconds()),
	}
	_, err := C.sched_setattr(C.pid_t(pid), &cAttr, C.uint(0))
	return err
}

// sched_getattr(2)
func GetAttr(pid int) (SchedAttr, error) {
	attr := SchedAttr{}
	cAttr := C.struct_sched_attr{}
	_, err := C.sched_getattr(C.pid_t(pid), &cAttr, C.uint(C.SCHED_ATTR_SIZE), C.uint(0))
	if err == nil {
		attr.Policy = SchedPolicy(cAttr.sched_policy)
		attr.Flags = SchedFlag(cAttr.sched_flags)
		attr.Nice = int32(cAttr.sched_nice)
		// not sure why we need this tail underscore: sched_priority_
		attr.Priority = uint32(cAttr.sched_priority_)
		attr.Runtime = time.Duration(cAttr.sched_runtime)
		attr.Deadline = time.Duration(cAttr.sched_deadline)
		attr.Period = time.Duration(cAttr.sched_period)
	}
	return attr, err
}

# libearthquake-ld-preload.so: injects random delay/fault to syscalls for flaky testcases


Usage:
    
    $ EQ_LD_PRELOAD='{"debug":"true", "max_sleep":"300ns"}' LD_PRELOAD=../../bin/libearthquake-ld-preload.so some_dynlinked_prog
	

Currently, `libearthquake-ld-preload.so` is (intentionally) self-orchestrated. i.e., it not depend on Earthquake Orchestrator.

Although we are considering implementing orchestration API client, self-orchestrated mode will be still kept for false-positive avoidance and interoperability.


TODO: support much more syscalls

Bug: some apps are not working due to [a bug of Go runtime](https://github.com/golang/go/issues/12465).


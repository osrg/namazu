# earthquake.stp: injects random delay/fault to syscalls for flaky testcases

Requirement: SystemTap

Usage:
    
    $ sudo stap earthquake.stp -g -G NET_FI_PERMIL=0 -G NET_MAX_DELAY=0 -G VFS_MAX_DELAY=0 -G VFS_FI_PERMIL=0 -G TARGET_EXECNAME=java
	

`earthquake.stp` is self-orchestrated. i.e., it not depend on Earthquake Orchestrator.

ATTENTION: SystemTap does NOT support interruptible sleeps.
Note that delays are non-interruptible.

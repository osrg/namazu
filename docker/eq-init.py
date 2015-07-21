#!/usr/bin/python
"""
Init script for Earthquake Docker Image (osrg/earthquake)
Supported Env vars:
 - EQ_DOCKER_PRIVILEGED
"""

import os
import prctl
import subprocess
import sys

def log(s):
    print 'INIT: %s' % s

def is_privileged_mode():
    has_env = os.getenv('EQ_DOCKER_PRIVILEGED')
    has_cap = prctl.cap_permitted.sys_admin
    if has_env and not has_cap:
        raise RuntimeError('EQ_DOCKER_PRIVILEGED is set, but SYS_ADMIN cap is missing')
    return has_env

def run_command_and_exit(l):
    log('Starting command: %s' % l)
    rc = subprocess.call(l)
    log('Exiting with status %d..(%s)' % (rc, l))
    sys.exit(rc)
    
def get_remaining_args():
    return sys.argv[1:]

if __name__ == '__main__':
    com = ['/bin/bash', '--login', '-i']
    if is_privileged_mode():
        log('Running with privileged mode. Enabling DinD, OVS, and Ryu')
        com = ['wrapdocker', '/init.dind-ovs-ryu.sh']
    else:
        log('Running without privileged mode. Please set EQ_DOCKER_PRIVILEGED if you want to use Ethernet Inspector')

    log('Earthquake is installed on /earthquake. Please refer to /earthquake/README.md')
    run_command_and_exit(com + get_remaining_args())

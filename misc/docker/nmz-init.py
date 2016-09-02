#!/usr/bin/python
"""
Init script for Namazu Docker Image (osrg/namazu)
Supported Env vars:
 - NMZ_DOCKER_PRIVILEGED
"""

import os
import prctl
import subprocess
import sys

def log(s):
    print 'INIT: %s' % s

def is_privileged_mode():
    has_env = os.getenv('NMZ_DOCKER_PRIVILEGED')
    has_cap = prctl.cap_permitted.sys_admin
    if has_env and not has_cap:
        raise RuntimeError('NMZ_DOCKER_PRIVILEGED is set, but SYS_ADMIN cap is missing')
    return has_env

def run_daemons(l):
    for elem in l:
        log('Starting daemon: %s' % elem)
        rc = subprocess.call(elem)
        if rc != 0:
            log('Exiting with status %d..(%s)' % (rc, elem))
            sys.exit(rc)

def run_command_and_exit(l):
    log('Starting command: %s' % l)
    rc = subprocess.call(l)
    log('Exiting with status %d..(%s)' % (rc, l))
    sys.exit(rc)
    
def get_remaining_args():
    return sys.argv[1:]

if __name__ == '__main__':
    daemons = [
        ['service', 'mongodb', 'start']
    ]
    run_daemons(daemons)
    com = ['/bin/bash', '--login', '-i']
    if is_privileged_mode():
        log('Running with privileged mode. Enabling DinD, OVS, and Ryu')
        com = ['wrapdocker', '/init.dind-ovs-ryu.sh']
    else:
        log('Running without privileged mode. Please set NMZ_DOCKER_PRIVILEGED if you want to use Ethernet Inspector')

    log('Namazu is installed on $GOPATH/src/github.com/osrg/namazu. Please refer to namazu/README.md')
    run_command_and_exit(com + get_remaining_args())


#!/bin/bash
set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

########## Boot ##########
if [ -z $EQ_DISABLE ]; then
    CHECK_PYTHONPATH
    START_SWITCH
    START_INSPECTOR
    # TODO: check boot failure
fi
START_DOCKER
SET_PIPEWORK
START_ETCD
#PAUSE
SLEEP ${ETCD_START_WAIT_SECS} # the user should increase this, if could not reproduce the bug

export ETCDCTL_PEERS="http://192.168.42.1:4001,http://192.168.42.2:4001,http://192.168.42.3:4001"
echo "writing k1"
etcdctl --no-sync --timeout 10s set /k1 v1 &
echo "result: $?"
echo "writing k2"
etcdctl --no-sync --timeout 10s set /k2 v2 &
echo "result: $?"
echo "writing k3"
etcdctl --no-sync --timeout 10s set /k3 v3 &
echo "result: $?"

etcdctl --no-sync --timeout 10s member add etcd4 http://192.168.42.4:4001 # dummy
exit 0

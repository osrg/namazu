#!/bin/bash
. ${NMZ_MATERIALS_DIR}/lib.sh

########## Boot ##########
if [ -z $NMZ_DISABLE ]; then
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
$NMZ_MATERIALS_DIR/etcd_testbed/etcd/bin/etcdctl --no-sync --timeout 10s set /k1 v1

exit 0

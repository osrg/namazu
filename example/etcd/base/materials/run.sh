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

exit 0

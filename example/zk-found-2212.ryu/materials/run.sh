#!/bin/bash
set -e # exit on an error
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
START_ZOOKEEPER
#PAUSE
SLEEP ${ZK_START_WAIT_SECS} # the user should increase this, if could not reproduce the bug
# NOTE: we don't finish running "run.sh" here, as we need to validate the living ensemble, not dead one.

########## Validation ##########
CHECK_FLE_STATES # see also validate.sh

########## Shutdown ##########
KILL_DOCKER
if [ -z $NMZ_DISABLE ]; then
    KILL_SWITCH
    KILL_INSPECTOR
fi

exit 0

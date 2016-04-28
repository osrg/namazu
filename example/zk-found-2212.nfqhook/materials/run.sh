#!/bin/bash
set -e # exit on an error
. ${NMZ_MATERIALS_DIR}/lib.sh

########## Boot ##########
CHECK_PREREQUISITES # already checked in init, but check again
INIT_ZOOKEEPER
if [ -z ${NMZ_DISABLE} ]; then
    INSTALL_IPTABLES_RULE
    START_NFQHOOK
    START_INSPECTOR
    # TODO: check boot failure
fi

START_ZOOKEEPER
#PAUSE
SLEEP ${ZK_START_WAIT_SECS} # the user should increase this, if could not reproduce the bug
# NOTE: we don't finish running "run.sh" here, as we need to validate the living ensemble, not dead one.

########## Validation ##########
CHECK_FLE_STATES # see also validate.sh
#PAUSE
########## Shutdown ##########
KILL_ZOOKEEPER
if [ -z ${NMZ_DISABLE} ]; then
    UNINSTALL_IPTABLES_RULE
    KILL_NFQHOOK
    KILL_INSPECTOR
fi

exit 0

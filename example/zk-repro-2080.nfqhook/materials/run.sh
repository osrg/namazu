#!/bin/bash
set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

########## Boot ##########
if [ -z ${EQ_DISABLE} ]; then
    CHECK_PYTHONPATH
    START_NFQHOOK
    START_INSPECTOR
    if [ -z ${SYSLOG_DISABLE} ]; then
      START_SYSLOG_INSPECTOR
    fi
    # TODO: check boot failure
fi
START_ZK_TEST
#PAUSE

########## Shutdown ##########
if [ -z ${EQ_DISABLE} ]; then
    KILL_NFQHOOK
    KILL_INSPECTOR
    if [ -z ${SYSLOG_DISABLE} ]; then
      KILL_SYSLOG_INSPECTOR
    fi
fi

exit 0

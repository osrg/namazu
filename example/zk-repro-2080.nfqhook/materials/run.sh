#!/bin/bash
set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

## some bad hack: allows modifying materials for interactive experiments
cp ${EQ_MATERIALS_DIR}/../config.json ${EQ_WORKING_DIR}/config.json.BAK
cp ${EQ_MATERIALS_DIR}/zk_inspector.py ${EQ_WORKING_DIR}/zk_inspector.py.BAK

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

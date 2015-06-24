#!/bin/bash
set -e # exit on an error
source lib.sh

EXP_DIR=/tmp/eq-run-experiment/$(date +"%Y%m%d.%H%M%S")
mkdir -p ${EXP_DIR}
IMPORTANT "===== EXPERIMENT BEGIN (${EXP_DIR}) ====="


BOOT 1
BOOT 2
INFO "Invoking reconfig 2"
RECONFIG_ADD_SERVER 2 10 5 #args: sid, trials, sleep_secs
ZKSYNC /
BOOT 3
INFO "Invoking reconfig 3"
RECONFIG_ADD_SERVER 3 5 5 || (IMPORTANT "PERHAPS ZOOKEEPER-2172 WAS REPRODUCED!"; touch ${EXP_DIR}/REPRODUCED)

if [ -z $DISABLE_EQ ]; then
    INFO "Stopping inspection"
    STOP_EQ_INSPECTION
    SLEEP 3
    last_exp=$(ls /tmp/eq/search/history | tail -1)
    cp -r /tmp/eq/search/history/${last_exp} ${EXP_DIR}
fi

INFO "Collecting ZooKeeper log files"
COLLECT_ZOOKEEPER_LOG ${EXP_DIR}

INFO "Trying to create znodes"
CREATE_ZNODE /foo bar || true

if [ -e ${EXP_DIR}/REPRODUCED ]; then
    IMPORTANT "PLEASE CHECK WHETHER ZOOKEEPER-2172 WAS REPRODUCED!!"
    IMPORTANT "DROPPING TO SHELL"
    bash --login -i
fi    


IMPORTANT "===== EXPERIMENT END (${EXP_DIR}) ====="

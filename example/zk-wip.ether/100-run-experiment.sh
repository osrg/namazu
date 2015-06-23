#!/bin/bash
set -e # exit on an error
source lib.sh

EXP_DIR=/tmp/eq-run-experiment/$(date +"%Y%m%d.%H%M%S")
mkdir -p ${EXP_DIR}
IMPORTANT "===== EXPERIMENT BEGIN (${EXP_DIR}) ====="

BOOT 1 0 #args: sid, sleep_secs
BOOT 2 0
INFO "Invoking reconfig 2"
RECONFIG_ADD_SERVER 2 10 5 #args: sid, trials, sleep_secs
[ $? -eq 0 ]
BOOT 3 5
INFO "Invoking reconfig 3"
RECONFIG_ADD_SERVER 3 100 5
[ $? -eq 0 ]
SLEEP 10

if [ -z $DISABLE_EQ ]; then
    INFO "Stopping inspection"
    STOP_EQ_INSPECTION
    SLEEP 3
    last_exp=$(ls /tmp/eq/search/history | tail -1)
    cp -r /tmp/eq/search/history/${last_exp} ${EXP_DIR}
fi


CREATE_ZNODE /foo bar

INFO "Collecting ZooKeeper log files"
COLLECT_ZOOKEEPER_LOG ${EXP_DIR}

INFO "Killing the docker containers"
KILL_DOCKERS

IMPORTANT "===== EXPERIMENT END (${EXP_DIR}) ====="

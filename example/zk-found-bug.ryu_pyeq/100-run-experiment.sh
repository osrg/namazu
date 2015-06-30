#!/bin/bash
set -e # exit on an error
source lib.sh

INFO "Cleaning veths"
CLEAN_VETHS

INFO "Creating docker containers"
START_DOCKER

INFO "Setting up pipework"
SET_PIPEWORK

INFO "Starting ZooKeeper"
START_ZOOKEEPER

INFO "Waiting for ${ZK_START_WAIT_SECS}"
sleep $ZK_START_WAIT_SECS

if [ -z $DISABLE_EQ ]; then
    INFO "Stopping inspection"
    STOP_EQ_INSPECTION
fi

INFO "Checking whether the bug was reproduced"
CHECK_BUG_REPRODUCED
IMPORTANT "The bug was NOT reproduced"

INFO "Collecting ZooKeeper log files"
COLLECT_ZOOKEEPER_LOG
INFO "Collected to ${ZOOKEEPER_LOG_DIR}."

INFO "Killing the docker container"
KILL_DOCKER

IMPORTANT "If the bug was not reproduced, please run this script again. You may have to change time_slice and time_bound in ${CONFIG_JSON}."

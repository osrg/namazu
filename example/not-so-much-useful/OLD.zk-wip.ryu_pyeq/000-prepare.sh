#!/bin/bash
set -e # exit on an error
source lib.sh

mkdir -p /tmp/eq
(docker images | grep $DOCKER_IMAGE_NAME) || \
    (
	INFO "Building docker image"
	BUILD_DOCKER_IMAGE
    )

(ls -1 zk_testbed/zookeeper/build | grep jar) || \
    (
	INFO "Building zookeeper"
	( cd zk_testbed/zookeeper; ant)
    )


if [ -z $DISABLE_EQ ]; then
    INFO "Starting Earthquake Ethernet Switch"
    EQ_SWITCH > /tmp/eq-switch.log 2>&1 &
    EQ_SWITCH_PID=$!
    echo $EQ_SWITCH_PID > /tmp/eq-switch.pid

    INFO "Starting Earthquake Orchestrator"
    EQ_ORCHESTRATOR > /tmp/eq-orchestrator.log 2>&1 &
    EQ_ORCHESTRATOR_PID=$!
    echo $EQ_ORCHESTRATOR_PID > /tmp/eq-orchestrator.pid

    INFO "Starting Earthquake Ethernet Inspector"
    EQ_INSPECTOR > /tmp/eq-inspector.log 2>&1 &
    EQ_INSPECTOR_PID=$!
    echo $EQ_INSPECTOR_PID > /tmp/eq-inspector.pid

    IMPORTANT "Please kill the processes (switch=${EQ_SWITCH_PID}, orchestrator=${EQ_ORCHESTRATOR_PID}, and inspector=${EQ_INSPECTOR_PID}) after you finished all of the experiments"
fi

IMPORTANT "Please continue to 100-run-experiment.sh.."

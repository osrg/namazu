#!/bin/bash
set -e # exit on an error
source lib.sh

(docker images | grep $DOCKER_IMAGE_NAME) || \
    (
	INFO "Building docker image"
	BUILD_DOCKER_IMAGE
    )

if [ -z $DISABLE_EQ ]; then
    INFO "Starting Earthquake Ethernet Switch"
    EQ_SWITCH > /tmp/eq-switch.log 2>&1 &
    EQ_SWITCH_PID=$!

    INFO "Starting Earthquake Orchestrator"
    EQ_ORCHESTRATOR > /tmp/eq-orchestrator.log 2>&1 &
    EQ_ORCHESTRATOR_PID=$!

    INFO "Starting Earthquake Ethernet Inspector"
    EQ_INSPECTOR > /tmp/eq-inspector.log 2>&1 &
    EQ_INSPECTOR_PID=$!

    IMPORTANT "Please kill the processes (switch=${EQ_SWITCH_PID}, orchestrator=${EQ_ORCHESTRATOR_PID}, and inspector=${EQ_INSPECTOR_PID}) after you finished all of the experiments"
fi

IMPORTANT "Please continue to 100-run-experiment.sh.."

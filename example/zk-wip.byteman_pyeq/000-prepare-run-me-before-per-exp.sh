#!/bin/bash
set -e # exit on an error
source lib.sh

mkdir -p /tmp/eq

(ls -1 zookeeper/build | grep jar) || \
    (
	INFO "Building zookeeper"
	( cd zookeeper; ant)
    )


if [ -z $DISABLE_EQ ]; then
    INFO "Starting Earthquake Orchestrator"
    EQ_ORCHESTRATOR > /tmp/eq-orchestrator.log 2>&1 &
    EQ_ORCHESTRATOR_PID=$!
    echo $EQ_ORCHESTRATOR_PID > /tmp/eq-orchestrator.pid

    IMPORTANT "Please kill the processes (orchestrator=${EQ_ORCHESTRATOR_PID}) after you finished all of the experiments"
fi

IMPORTANT "Please continue to 100-run-experiment.sh.."

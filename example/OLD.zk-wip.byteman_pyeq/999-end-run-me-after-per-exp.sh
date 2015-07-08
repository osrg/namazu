#!/bin/bash
source lib.sh

SHUTDOWN 1
SHUTDOWN 2
SHUTDOWN 3

if [ -z $DISABLE_EQ ]; then
    INFO "Killing Earthquake Orchestrator"
    kill -9 $(cat /tmp/eq-orchestrator.pid)
fi

#!/bin/bash
source lib.sh

if [ -z $DISABLE_EQ ]; then
    INFO "Killing Earthquake Ethernet Inspector"
    kill -9 $(cat /tmp/eq-inspector.pid)

    INFO "Killing Earthquake Orchestrator"
    kill -9 $(cat /tmp/eq-orchestrator.pid)
fi

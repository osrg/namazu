#!/bin/bash

## CONFIG
TEST_USER=nfqhooked
TEST_PWD=$(pwd)/zookeeper
TEST_COMMAND="ant -Dtest.output=yes -Dtestcase=ReconfigRecoveryTest -Dtest.method=testCurrentObserverIsParticipantInNewConfig test-core-java"
CONFIG_JSON=config.json

## ENVS
EQDIR=$(pwd)/../..
export LD_LIBRARY_PATH=$EQDIR/bin:$LD_LIBRARY_PATH
export PYTHONPATH=.:$EQDIR:$PYTHONPATH

## FUNCS
function INFO(){
    echo -e "\e[104m\e[97m[INFO]\e[49m\e[39m $@"
}

function IMPORTANT(){
    echo -e "\e[105m\e[97m[IMPORTANT]\e[49m\e[39m $@"
}

function SLEEP(){
    echo -n $(INFO "Sleeping(${1} secs)..")
    sleep ${1}
    echo "Done"
}

function DO_TEST(){
    TEST_LOGFILE=$(mktemp)
    IMPORTANT "--- TEST BEGIN: ${TEST_COMMAND} ---"
    sudo -u ${TEST_USER} bash -c "( cd ${TEST_PWD}; ${TEST_COMMAND} )" 2>&1 | tee ${TEST_LOGFILE}
    IMPORTANT "--- TEST DONE: ${TEST_COMMAND} ---"
    grep FAILED ${TEST_LOGFILE} && return $(false)
    return $(true)
}

function STOP_EQ_INSPECTION(){
    curl http://localhost:10000/ctrl_api/v1/forcibly_inspection_end
}

shopt -s expand_aliases
alias EQ_NFQHOOK="sudo PYTHONPATH=${PYTHONPATH} python ./zk_nfqhook.py"

alias EQ_ORCHESTRATOR="python -m pyearthquake.cmd.orchestrator_loader $CONFIG_JSON"

alias EQ_INSPECTOR="python ./dumb_inspector.py"

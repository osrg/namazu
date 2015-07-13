#!/bin/bash

## CONFIG
DOCKER_IMAGE_NAME=zk_testbed
CONFIG_JSON=config.json
ZK_START_WAIT_SECS=5

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


function BUILD_DOCKER_IMAGE(){
    docker build -t $DOCKER_IMAGE_NAME zk_testbed
}

shopt -s expand_aliases
alias EQ_SWITCH="ryu-manager ./zk_switch.py"

alias EQ_ORCHESTRATOR="python -m pyearthquake.cmd.orchestrator_loader $CONFIG_JSON"

alias EQ_INSPECTOR="python ./zk_inspector.py"


function CLEAN_VETHS(){
    garbages=$(ip a | egrep -o 'veth.*:' | sed -e s/://g)
    for f in $garbages; do sudo ip link delete $f; done
}

function START_DOCKER(){
    for f in $(seq 1 3); do
        docker run -i -t -d -e ZKID=${f} -e ZKENSEMBLE=1 -h zk${f} --name zk${f} $DOCKER_IMAGE_NAME /bin/bash;
    done
}

function SET_PIPEWORK(){
    for f in $(seq 1 3); do sudo pipework ovsbr0 zk${f} 192.168.42.${f}/24; done
}

function START_ZOOKEEPER(){
    for f in $(seq 1 3); do docker exec -d zk${f} /bin/bash -c '/init.py > /log 2>&1'; done
}

function STOP_EQ_INSPECTION(){
    curl --data {} http://localhost:10080/api/v2/ctrl/force_terminate
}

function CHECK_BUG_REPRODUCED(){
    ./check.py || (IMPORTANT "THE BUG WAS REPRODUCED!"; false)
}    

function COLLECT_ZOOKEEPER_LOG(){
    mkdir -p /tmp/eq-zklog
    ZOOKEEPER_LOG_DIR=$(mktemp -d /tmp/eq-zklog/XXXXX)
    for f in zk1 zk2 zk3; do docker cp $f:/log $ZOOKEEPER_LOG_DIR/$f; done
}

function KILL_DOCKER(){
    docker rm -f zk1 zk2 zk3
}

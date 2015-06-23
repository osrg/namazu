#!/bin/bash

## CONFIG
DOCKER_IMAGE_NAME=zk_wip_testbed
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


function BUILD_DOCKER_IMAGE(){
    docker build -t $DOCKER_IMAGE_NAME zk_testbed
}

shopt -s expand_aliases
alias EQ_SWITCH="ryu-manager ./zk_switch.py"

alias EQ_ORCHESTRATOR="python -m pyearthquake.cmd.orchestrator_loader $CONFIG_JSON"

alias EQ_INSPECTOR="python ./dumb_inspector.py"

function PREBOOT(){
    sid=$1
    wait_secs=$2
    INFO "=== Pre-Boot ${sid} ==="
    INFO "Pre-Boot ${sid}: Starting Docker"
    START_DOCKER ${sid}
    INFO "Pre-Boot ${sid}: Setting pipework"    
    SET_PIPEWORK ${sid}
    SLEEP ${wait_secs}
    INFO "Pre-Boot ${sid}: Pinging"        
    ping -c 3 192.168.42.${sid} || (return $(false))
}

function BOOT(){
    sid=$1
    wait_secs=$2
    INFO "=== Boot ${sid} ==="
    INFO "Boot ${sid}: Starting ZooKeeper"        
    START_ZOOKEEPER ${sid}
    SLEEP ${wait_secs}
}


function SLEEP(){
    echo -n $(INFO "Sleeping(${1} secs)..")
    sleep ${1}
    echo "Done"
}    

function START_DOCKER(){
    sid=$1
    docker run -i -t -d -e ZKID=${sid} -e ZKENSEMBLE=1 -h zk${sid} --name zk${sid} $DOCKER_IMAGE_NAME /bin/bash
}

function SET_PIPEWORK(){
    sid=$1
    sudo pipework ovsbr0 zk${sid} 192.168.42.${sid}/24
}

function START_ZOOKEEPER(){
    sid=$1
    docker exec -d zk${sid} /bin/bash -c '/init.py > /log 2>&1'
}

function STOP_EQ_INSPECTION(){
    for f in InspectionEndEvents/*.json; do curl --data @$f http://localhost:10000/api/v1; done    
}

function CHECK_BUG_REPRODUCED(){
    IMPORTANT "Please check bug reproduced or not, by your self, then press ret"
    read
}    

function COLLECT_ZOOKEEPER_LOG(){
    d=$1
    for f in zk1 zk2 zk3; do docker cp $f:/log $d/$f; done
}

function KILL_DOCKERS(){
    docker rm -f zk1 zk2 zk3
}

ZKCLI=zk_testbed/zookeeper/bin/zkCli.sh 
ZKCLI_ARG="-server 192.168.42.1"
function ZKSYNC(){
    $ZKCLI $ZKCLI_ARG sync $1
}
function RECONFIG_ADD_SERVER(){
    sid=$1
    trials=$2
    sleep_secs=$3
    for f in $(seq 1 ${trials}); do
	INFO "Reconfig (sid=${sid}, trial=${f} of ${trials})"
	tmp=$(mktemp)
	$ZKCLI $ZKCLI_ARG reconfig -add server.${sid}=192.168.42.${sid}:2888:3888:participant\;2181 2>&1 | tee ${tmp}
	errors=$(grep KeeperErrorCode ${tmp} | wc -l)
	if [ $errors -eq 0 ]; then
	    INFO "Reconfig success (sid=${sid})"; rm -f ${tmp}; return $(true)
	fi
	INFO "Reconfig fail (sid=${sid}, trial=${f} of ${trials})"	
	rm -f $tmp
	SLEEP ${sleep_secs}
    done
    return $(false)
}

function CREATE_ZNODE(){
    $ZKCLI $ZKCLI_ARG create $1 $2
}

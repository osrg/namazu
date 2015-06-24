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

alias EQ_ORCHESTRATOR="python -m pyearthquake.cmd.orchestrator_loader $CONFIG_JSON"

AGENT_CP=$(pwd)/../../bin/earthquake-inspector.jar

function BOOT(){
    sid=$1
    sdir=/tmp/eq/zk/${sid}/
    rm -rf ${sdir}
    mkdir -p ${sdir}
    echo ${sid} > ${sdir}/myid

    rm -rf /tmp/eq/zk-conf
    cp -r zk-conf /tmp/eq

    export EQ_ENV_PROCESS_ID=zk${sid}
    export SERVER_JVMFLAGS="-javaagent:${AGENT_CP}=script:zk.btm" 
    export ZOO_LOG_DIR=/tmp/eq/zk-log/${sid}/ 
    mkdir -p $ZOO_LOG_DIR
    

    zookeeper/bin/zkServer.sh --config /tmp/eq/zk-conf/${sid} start    
}    

function SHUTDOWN(){
    sid=$1
    zookeeper/bin/zkServer.sh --config /tmp/eq/zk-conf/${sid} stop
}    


function SLEEP(){
    echo -n $(INFO "Sleeping(${1} secs)..")
    sleep ${1}
    echo "Done"
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
    cp -r /tmp/eq/zk-log $d
}


ZKCLI=zookeeper/bin/zkCli.sh 
ZKCLI_ARG="-server 127.0.0.1"
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
	$ZKCLI $ZKCLI_ARG reconfig -add server.${sid}=127.0.0.1:$((2887+sid)):$((3887+sid)):participant\;$((2180+sid)) 2>&1 | tee ${tmp}
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

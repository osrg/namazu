#!/bin/bash

## CONFIG
# EQ_DISABLE=1 # set to disable earthquake
ETCD_GIT_COMMIT=${ETCD_GIT_COMMIT:-master}
DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME:-etcd_testbed}

ETCD_START_WAIT_SECS=${ETCD_START_WAIT_SECS:-10}
PAUSE_ON_FAILURE=${PAUSE_ON_FAILURE:-0}

## GENERIC FUNCS
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

function PAUSE(){
    TMP=$(mktemp)
    IMPORTANT "PAUSING. remove ${TMP} to continue"
    while [ -e $TMP ]; do
      sleep 3
    done
}

## FUNCS (INIT)
function CHECK_PREREQUISITES() {
    INFO "Checking whether Docker is installed"
    hash docker
    INFO "Checking whether pipework is installed"
    hash pipework
    INFO "Checking whether ryu is installed"
    hash ryu-manager
    INFO "Checking whether hookswitch is installed"
    hash hookswitch-of13
    INFO "Checking whether ovsbr0 is configured as 192.168.42.254"
    ip addr show ovsbr0
    test X$(ip addr show ovsbr0 | sed -nEe 's/^[ \t]*inet[ \t]*([0-9.]+)\/.*$/\1/p') = X192.168.42.254
}

function FETCH_ETCD() {
    ( cd ${EQ_MATERIALS_DIR}/etcd_testbed;
      INFO "Fetching etcd"
      git clone https://github.com/coreos/etcd.git
      INFO "Checking out etcd@${ETCD_GIT_COMMIT}"
      INFO "You can change the etcd version by setting ETCD_GIT_COMMIT"
      cd etcd
      git checkout ${ETCD_GIT_COMMIT}
      ./build )
}

function BUILD_DOCKER_IMAGE() {
    ( cd ${EQ_MATERIALS_DIR}/etcd_testbed;
      docker_build_log=${EQ_MATERIALS_DIR}/docker-build.log
      INFO "Building Docker Image ${DOCKER_IMAGE_NAME} (${docker_build_log})";
      docker build -t ${DOCKER_IMAGE_NAME} . > ${docker_build_log} )
}


## FUNCS (BOOT)
export EQ_ETHER_ZMQ_ADDR="ipc://${EQ_WORKING_DIR}/ether_inspector"

function CHECK_PYTHONPATH() {
    INFO "Checking PYTHONPATH(=${PYTHONPATH})"
    ## used for etcd_inspector
    python -c "import pyearthquake"
}    

function START_SWITCH() {
    INFO "Starting HookSwitch"
    hookswitch-of13 ${EQ_ETHER_ZMQ_ADDR} --debug --tcp-ports=4001,7001 > ${EQ_WORKING_DIR}/switch.log 2>&1 &
    pid=$!
    INFO "Switch PID: ${pid}"
    echo ${pid} > ${EQ_WORKING_DIR}/switch.pid
}

function START_INSPECTOR() {
    INFO "Starting Earthquake Ethernet Inspector"
    python ${EQ_MATERIALS_DIR}/etcd_inspector.py > ${EQ_WORKING_DIR}/inspector.log 2>&1 &
    pid=$!
    INFO "Inspector PID: ${pid}"
    echo ${pid} > ${EQ_WORKING_DIR}/inspector.pid
}

function START_DOCKER() {
    for f in $(seq 1 3); do
	    INFO "Starting Docker container etcd${f} from ${DOCKER_IMAGE_NAME}"
      docker run -i -t -d -e ETCDID=${f} -h etcd${f} --name etcd${f} ${DOCKER_IMAGE_NAME} /bin/bash;
    done
}

function SET_PIPEWORK() {
    for f in $(seq 1 3); do
	    INFO "Assigning 192.168.42.${f}/24 (ovsbr0) to etcd${f}"
	    pipework ovsbr0 etcd${f} 192.168.42.${f}/24;
    done
}

function START_ETCD() {
    for f in $(seq 1 3); do 
	  INFO "Starting etcd(id: ${f}) in Docker container etcd${f}"
	  docker exec -d etcd${f} /bin/bash -c 'bash /init.sh > /log 2>&1';
    done
}

## FUNCS (VALIDATION)

## FUNCS (SHUTDOWN)
function KILL_SWITCH() {
    pid=$(cat ${EQ_WORKING_DIR}/switch.pid)
    INFO "Killing Switch, PID: ${pid}"
    kill -9 ${pid}
}

function KILL_INSPECTOR() {
    pid=$(cat ${EQ_WORKING_DIR}/inspector.pid)
    INFO "Killing Inspector, PID: ${pid}"
    kill -9 ${pid}
}

function KILL_DOCKER() {
    docker stop etcd1 etcd2 etcd3
    for f in $(seq 1 3); do
	   INFO "Killing Docker container etcd${f} (log:${EQ_WORKING_DIR}/etcd${f})"
	   mkdir ${EQ_WORKING_DIR}/etcd${f}
	   docker cp etcd${f}:/log ${EQ_WORKING_DIR}/etcd${f}
	   docker rm -f etcd${f}
    done
}

function CLEAN_VETHS(){
    INFO "Removing garbage veths"
    IMPORTANT "CLEAN_VETHS() has not been tested well"
    garbages=$(ip a | egrep -o 'veth.*:' | sed -e s/://g)
    for f in $garbages; do ip link delete $f; done
}

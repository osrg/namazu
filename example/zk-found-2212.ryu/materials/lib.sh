#!/bin/bash

## CONFIG
# NMZ_DISABLE=1 # set to disable namazu
## ZooKeeper maintainers seems switched git repo and the previous 98a3c is now available as 02d15 on the master
# ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-98a3cabfa279833b81908d72f1c10ee9f598a045} #(Tue Jun 2 19:17:09 2015 +0000)
ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-02d1505e4df8c8669b89b74be37aa3a1025422ab} #(Tue Jun 2 19:17:09 2015 +0000)
DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME:-zk_testbed}

ZK_START_WAIT_SECS=${ZK_START_WAIT_SECS:-10}
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
    INFO "Checking whether zktraffic is installed"
    python -c "from zktraffic.omni.omni_sniffer import OmniSniffer"
}

function FETCH_ZK() {
    ( cd ${NMZ_MATERIALS_DIR}/zk_testbed;
      INFO "Fetching ZooKeeper"
      git clone https://github.com/apache/zookeeper.git
      INFO "Checking out ZooKeeper@${ZK_GIT_COMMIT}"
      INFO "You can change the ZooKeeper version by setting ZK_GIT_COMMIT"
      cd zookeeper
      git checkout ${ZK_GIT_COMMIT} )
}

function BUILD_DOCKER_IMAGE() {
    ( cd ${NMZ_MATERIALS_DIR}/zk_testbed;
      docker_build_log=${NMZ_MATERIALS_DIR}/docker-build.log
      INFO "Building Docker Image ${DOCKER_IMAGE_NAME} (${docker_build_log})"
      docker build -t ${DOCKER_IMAGE_NAME} . > ${docker_build_log} )
}


## FUNCS (BOOT)
export NMZ_ETHER_ZMQ_ADDR="ipc://${NMZ_WORKING_DIR}/ether_inspector"

function CHECK_PYTHONPATH() {
    INFO "Checking PYTHONPATH(=${PYTHONPATH})"
    ## used for zk_inspector
    python -c "import pynmz"
}    

function START_SWITCH() {
    INFO "Starting HookSwitch"
    hookswitch-of13 ${NMZ_ETHER_ZMQ_ADDR} --debug --tcp-ports=2888,3888 > ${NMZ_WORKING_DIR}/switch.log 2>&1 &
    pid=$!
    INFO "Switch PID: ${pid}"
    echo ${pid} > ${NMZ_WORKING_DIR}/switch.pid
}

function START_INSPECTOR() {
    INFO "Starting Namazu Ethernet Inspector"
    python ${NMZ_MATERIALS_DIR}/zk_inspector.py > ${NMZ_WORKING_DIR}/inspector.log 2>&1 &
    pid=$!
    INFO "Inspector PID: ${pid}"
    echo ${pid} > ${NMZ_WORKING_DIR}/inspector.pid
}

function START_DOCKER() {
    for f in $(seq 1 3); do
	INFO "Starting Docker container zk${f} from ${DOCKER_IMAGE_NAME}"
	docker run -i -t -d -e ZKID=${f} -e ZKENSEMBLE=1 -h zk${f} --name zk${f} ${DOCKER_IMAGE_NAME} /bin/bash
    done
}

function SET_PIPEWORK() {
    for f in $(seq 1 3); do
	INFO "Assigning 192.168.42.${f}/24 (ovsbr0) to zk${f}"
	pipework ovsbr0 zk${f} 192.168.42.${f}/24
    done
}

function START_ZOOKEEPER() {
    for f in $(seq 1 3); do 
	INFO "Starting ZooKeeper(sid=${f}) in Docker container zk${f}"
	docker exec -d zk${f} python /init.py
    done
}

## FUNCS (VALIDATION)

function CHECK_FLE_STATES() {
    INFO "Checking FLE states"
    result=0
    (python ${NMZ_MATERIALS_DIR}/check-fle-states.py > ${NMZ_WORKING_DIR}/check-fle-states.log) || result=$?
    echo ${result} > ${NMZ_WORKING_DIR}/check-fle-states.result
    if [ ${result} != 0 ]; then
	IMPORTANT "Failure: ${result} (${NMZ_WORKING_DIR}/check-fle-states.log)"
	if [ ${PAUSE_ON_FAILURE} != 0 ]; then
	    IMPORTANT "Pausing.. please check whether this is a false-positive or not"
	    PAUSE
	fi
	# do not return $(false) here
    fi
}

## FUNCS (SHUTDOWN)
function KILL_SWITCH() {
    pid=$(cat ${NMZ_WORKING_DIR}/switch.pid)
    INFO "Killing Switch, PID: ${pid}"
    kill -9 ${pid}
}

function KILL_INSPECTOR() {
    pid=$(cat ${NMZ_WORKING_DIR}/inspector.pid)
    INFO "Killing Inspector, PID: ${pid}"
    kill -9 ${pid}
}

function KILL_DOCKER() {
    for f in $(seq 1 3); do
	INFO "Killing Docker container zk${f} (log:${NMZ_WORKING_DIR}/zk${f})"
	docker exec zk${f} /zk/bin/zkServer.sh stop
	mkdir ${NMZ_WORKING_DIR}/zk${f}
        docker cp zk${f}:/zk/jacoco.exec ${NMZ_WORKING_DIR}/zk${f}
	docker cp zk${f}:/zk/logs ${NMZ_WORKING_DIR}/zk${f}
	docker rm -f zk${f}
    done
}

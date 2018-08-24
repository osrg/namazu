#!/bin/bash

## CONFIG
# NMZ_DISABLE=1 # set to disable namazu
## ZooKeeper maintainers seems switched git repo and the previous 98a3c is now available as 02d15 on the master
# ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-98a3cabfa279833b81908d72f1c10ee9f598a045} #(Tue Jun 2 19:17:09 2015 +0000)
ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-02d1505e4df8c8669b89b74be37aa3a1025422ab} #(Tue Jun 2 19:17:09 2015 +0000)
ZK_START_WAIT_SECS=${ZK_START_WAIT_SECS:-10}
PAUSE_ON_FAILURE=${PAUSE_ON_FAILURE:-0}
NFQ_USER=${NFQ_USER:-nfqhooked}
NFQ_NUMBER=${NFQ_NUMBER:-42}

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
    INFO "Checking whether JDK is installed"
    hash javac
    
    INFO "Checking whether ant is installed"
    hash ant

    INFO "Checking whether unzip is installed"
    hash unzip

    INFO "Checking whether erb is installed"
    hash erb

    INFO "Checking whether zktraffic is installed"
    python -c "from zktraffic.omni.omni_sniffer import OmniSniffer"
    
    INFO "Checking whether hookswitch is installed"
    hash hookswitch-nfq

    if [ -f /proc/sys/net/ipv4/tcp_autocorking ]; then
	INFO "Checking whether tcp_autocorking (introduced in Linux 3.14) is disabled"
	test $(cat /proc/sys/net/ipv4/tcp_autocorking) = 0
    fi
    
    INFO "Checking existence of user \"${NFQ_USER}\""
    id -u ${NFQ_USER}

    INFO "Checking whether NFQUEUE ${NFQ_NUMBER} is available"
    test $(iptables -n -L -v | grep "NFQUEUE num ${NFQ_NUMBER}" | wc -l) = 0

    INFO "Checking PYTHONPATH"
    ## used for zk_inspector
    python -c "import pynmz"
    
}

function FETCH_ZK() {
    ( cd ${NMZ_MATERIALS_DIR};
      INFO "Fetching ZooKeeper"
      if [ -z ${ZK_SOURCE_DIR} ]; then
	  git clone https://github.com/apache/zookeeper.git
	  INFO "Checking out ZooKeeper@${ZK_GIT_COMMIT}"
	  INFO "You can change the ZooKeeper version by setting ZK_GIT_COMMIT"
	  cd zookeeper
	  git checkout ${ZK_GIT_COMMIT}
      else
	  INFO "Copying from ${ZK_SOURCE_DIR}"
	  cp -R ${ZK_SOURCE_DIR} .
	  cd zookeeper
	  ant clean
      fi
    )
}

function FETCH_JACOCO() {
    INFO "Fetching JaCoCo"
    mkdir ${NMZ_MATERIALS_DIR}/jacoco
    (
	cd ${NMZ_MATERIALS_DIR}/jacoco
	curl -L -O http://search.maven.org/remotecontent?filepath=org/jacoco/jacoco/0.7.5.201505241946/jacoco-0.7.5.201505241946.zip
	unzip -q jacoco-0.7.5.201505241946.zip
    )
}

function BUILD_ZK() {
    (
	cd ${NMZ_MATERIALS_DIR}/zookeeper
	INFO "Building ZooKeeper"
	ant
    )
}

## FUNCS (BOOT)
export NMZ_ETHER_ZMQ_ADDR="ipc://${NMZ_WORKING_DIR}/ether_inspector"

function INSTALL_IPTABLES_RULE() {
    INFO "Installing iptables rule for user=${NFQ_USER}, nfqueue=${NFQ_NUMBER}"
    iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u ${NFQ_USER}) -j NFQUEUE --queue-num ${NFQ_NUMBER}
}

function START_NFQHOOK() {
    INFO "Starting NFQ HookSwitch"
    hookswitch-nfq --nfq-number ${NFQ_NUMBER} --debug ${NMZ_ETHER_ZMQ_ADDR} > ${NMZ_WORKING_DIR}/nfqhook.log 2>&1 &
    pid=$!
    INFO "NFQ HookSwitch PID: ${pid}"
    echo ${pid} > ${NMZ_WORKING_DIR}/nfqhook.pid
}

function START_INSPECTOR() {
    INFO "Starting Namazu Ethernet Inspector"
    python ${NMZ_MATERIALS_DIR}/zk_inspector.py > ${NMZ_WORKING_DIR}/inspector.log 2>&1 &
    pid=$!
    INFO "Inspector PID: ${pid}"
    echo ${pid} > ${NMZ_WORKING_DIR}/inspector.pid
}

function INIT_ZOOKEEPER() {
    for myid in $(seq 1 3); do
	w=${NMZ_WORKING_DIR}/zk${myid}
	INFO "Initializing ZooKeeper(myid=${myid} at ${w})"
	mkdir -p ${w}
	(echo "<% myzkdir=\"${w}\"; myid=${myid} %>" && cat ${NMZ_MATERIALS_DIR}/zoo.cfg.erb) | erb > ${w}/zoo.cfg
	(echo "<% myzkdir=\"${w}\" %>" && cat ${NMZ_MATERIALS_DIR}/log4j.properties.erb) | erb > ${w}/log4j.properties
	chown -R ${NFQ_USER} ${w}
	sudo -E -u ${NFQ_USER} ${NMZ_MATERIALS_DIR}/zookeeper/bin/zkServer-initialize.sh --configfile=${w}/zoo.cfg --myid=${myid}
    done
}

function START_ZOOKEEPER() {
    for myid in $(seq 1 3); do
	w=${NMZ_WORKING_DIR}/zk${myid}
	INFO "Starting ZooKeeper(myid=${myid} at ${w})"
	jvmflags=-javaagent:${NMZ_MATERIALS_DIR}/jacoco/lib/jacocoagent.jar=destfile=${w}/jacoco.exec
	# this & is important
	ZOO_LOG_DIR=${w} JVMFLAGS=${jvmflags} sudo -E -u ${NFQ_USER} ${NMZ_MATERIALS_DIR}/zookeeper/bin/zkServer.sh --config ${w} start > /dev/null &
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
function KILL_ZOOKEEPER() {
    for myid in $(seq 1 3); do
	w=${NMZ_WORKING_DIR}/zk${myid}
	INFO "Stopping ZooKeeper(myid=${myid} at ${w})"	
	ZOO_LOG_DIR=${w} sudo -E -u ${NFQ_USER} ${NMZ_MATERIALS_DIR}/zookeeper/bin/zkServer.sh --config ${w} stop
    done
}

function UNINSTALL_IPTABLES_RULE() {
    INFO "Uninstalling iptables rule for user=${NFQ_USER}, nfqueue=${NFQ_NUMBER}"
    iptables -D OUTPUT -p tcp -m owner --uid-owner $(id -u ${NFQ_USER}) -j NFQUEUE --queue-num ${NFQ_NUMBER}
}

function KILL_NFQHOOK() {
    pid=$(cat ${NMZ_WORKING_DIR}/nfqhook.pid)
    INFO "Killing NFQHook, PID: ${pid}"
    kill -9 ${pid}
}

function KILL_INSPECTOR() {
    pid=$(cat ${NMZ_WORKING_DIR}/inspector.pid)
    INFO "Killing Inspector, PID: ${pid}"
    kill -9 ${pid}
}

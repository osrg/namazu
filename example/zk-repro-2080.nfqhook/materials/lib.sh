#!/bin/bash

## CONFIG
# EQ_DISABLE=1 # set to disable earthquake
ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-32fc1417dc649f8a3bb32a224eb6cca3181eb39f} #(Fri Jul 10 06:19:22 2015 +0000) # TODO: support other local ZK tree
ZK_TEST_COMMAND=${ZK_TEST_COMMAND:-ant -Dtestcase=ReconfigRecoveryTest -Dtest.method=testCurrentObserverIsParticipantInNewConfig -Dtest.output=true test-core-java}


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
    INFO "Checking whether Ant is installed"
    hash ant
    INFO "Checking whether zktraffic is installed"
    python -c "from zktraffic.omni.omni_sniffer import OmniSniffer"
    if [ -f /proc/sys/net/ipv4/tcp_autocorking ]; then
	INFO "Checking whether tcp_autocorking (introduced in Linux 3.14) is disabled"
	test $(cat /proc/sys/net/ipv4/tcp_autocorking) = 0
    fi
    INFO "Checking existence of user \"nfqhooked\""
    id -u nfqhooked
    if [ -z $EQ_DISABLE ]; then
	INFO "Checking whether NFQUEUE 42 is configured for a user \"nfqhooked\""
	iptables -n -L -v | grep "owner UID match $(id -u nfqhooked) NFQUEUE num 42"
    fi
}

function FETCH_ZK() {
    ( cd ${EQ_MATERIALS_DIR};
      INFO "Fetching ZooKeeper"
      git clone https://github.com/apache/zookeeper.git
      INFO "Checking out ZooKeeper@${ZK_GIT_COMMIT}"
      INFO "You can change the ZooKeeper version by setting ZK_GIT_COMMIT"
      cd zookeeper
      git checkout ${ZK_GIT_COMMIT}
    )
}

function BUILD_ZK() {
    ( cd ${EQ_MATERIALS_DIR}/zookeeper;
      INFO "Building ZooKeeper"
      ant
      ant test-init
      chown -R nfqhooked .
    )
}


## FUNCS (BOOT)
export EQ_ETHER_ZMQ_ADDR="ipc://${EQ_WORKING_DIR}/ether_inspector"

function CHECK_PYTHONPATH() {
    INFO "Checking PYTHONPATH(=${PYTHONPATH})"
    ## used for zk_nfqhook and zk_inspector
    python -c "import pyearthquake"
}    

function START_NFQHOOK() {
    INFO "Starting Earthquake Ethernet NFQHook"
    python ${EQ_MATERIALS_DIR}/zk_nfqhook.py > ${EQ_WORKING_DIR}/nfqhook.log 2>&1 &
    pid=$!
    INFO "NFQHook PID: ${pid}"
    echo ${pid} > ${EQ_WORKING_DIR}/nfqhook.pid
}

function START_INSPECTOR() {
    INFO "Starting Earthquake Ethernet Inspector"
    python ${EQ_MATERIALS_DIR}/zk_inspector.py > ${EQ_WORKING_DIR}/inspector.log 2>&1 &
    pid=$!
    INFO "Inspector PID: ${pid}"
    echo ${pid} > ${EQ_WORKING_DIR}/inspector.pid
}

function START_ZK_TEST() {
    INFO "Starting ZooKeeper testing (${ZK_TEST_COMMAND})"
    result=0
    if [ -z $EQ_DISABLE ]; then    
	(cd ${EQ_MATERIALS_DIR}/zookeeper; sudo -E -u nfqhooked sh -c "${ZK_TEST_COMMAND}" 2>&1 | tee ${EQ_WORKING_DIR}/zk-test.log) || result=$?
    else
	(cd ${EQ_MATERIALS_DIR}/zookeeper; sh -c "${ZK_TEST_COMMAND}" 2>&1 | tee ${EQ_WORKING_DIR}/zk-test.log) || result=$?
    fi
    echo ${result} > ${EQ_WORKING_DIR}/zk-test.result
}

## FUNCS (SHUTDOWN)
function KILL_NFQHOOK() {
    pid=$(cat ${EQ_WORKING_DIR}/nfqhook.pid)
    INFO "Killing NFQHook, PID: ${pid}"
    kill -9 ${pid}
}

function KILL_INSPECTOR() {
    pid=$(cat ${EQ_WORKING_DIR}/inspector.pid)
    INFO "Killing Inspector, PID: ${pid}"
    kill -9 ${pid}
}

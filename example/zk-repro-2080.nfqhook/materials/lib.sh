#!/bin/bash

## CONFIG
# NMZ_DISABLE=1 # set to disable namazu
ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-35e45512b5602eddbde9bb4ca0ef118d2fed7464} #(Sep 11, 2015)
# ZK_SKIP_JACOCO_PATCH=1 # set to skip applying ZOOKEEPER-2266-v2.patch (required only if already applied)
ZK_TEST_COMMAND=${ZK_TEST_COMMAND:-ant -Dtestcase=ReconfigRecoveryTest -Dtest.method=testCurrentObserverIsParticipantInNewConfig -Dtest.output=true test-core-java}
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

    INFO "Checking whether Ant is installed"
    hash ant

    INFO "Checking whether HookSwitch is installed"
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

function BUILD_ZK() {
    ( cd ${NMZ_MATERIALS_DIR}/zookeeper;
      INFO "Building ZooKeeper"
      cp -f ${NMZ_MATERIALS_DIR}/log4j.properties conf
      if [ -z ${ZK_SKIP_JACOCO_PATCH} ]; then
	  patch -p0 < ${NMZ_MATERIALS_DIR}/ZOOKEEPER-2266-v2.patch
      fi
      ant
      ant test-init
      chown -R ${NFQ_USER} .
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

function START_ZK_TEST() {
    INFO "Starting ZooKeeper testing (${ZK_TEST_COMMAND})"
    rm -rf ${NMZ_MATERIALS_DIR}/zookeeper/build/test/jacoco # this is important to avoid unexpected merging
    (cd ${NMZ_MATERIALS_DIR}/zookeeper; sudo -E -u ${NFQ_USER} sh -c "${ZK_TEST_COMMAND}" 2>&1 | tee ${NMZ_WORKING_DIR}/zk-test.log)
}

function COLLECT_COVERAGE() {
    ( cd ${NMZ_MATERIALS_DIR}/zookeeper; ant jacoco-report || INFO "no jacoco support.. (see ZOOKEEPER-2266)" )
    cp -r ${NMZ_MATERIALS_DIR}/zookeeper/build/test/jacoco ${NMZ_WORKING_DIR} || true
}

## FUNCS (SHUTDOWN)
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

function KILL_SYSLOG_INSPECTOR() {
    pid=$(cat ${NMZ_WORKING_DIR}/syslog_inspector.pid)
    INFO "Killing Syslog Inspector, PID: ${pid}"
    kill -9 ${pid}
}

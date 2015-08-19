#!/bin/bash

## CONFIG
# EQ_DISABLE=1 # set to disable earthquake
ZK_GIT_COMMIT=${ZK_GIT_COMMIT:-7f10f9c53a48b296dd27c7a104b303b10b045a89} #(Aug 27, 2015)

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
}

function FETCH_ZK() {
    ( cd ${EQ_MATERIALS_DIR};
      INFO "Fetching ZooKeeper"
      if [ -z $ZK_SOURCE_DIR ]; then
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
    ( cd ${EQ_MATERIALS_DIR}/zookeeper;
      INFO "Building ZooKeeper"
      ant
    )
    cp ${EQ_MATERIALS_DIR}/log4j.properties ${EQ_MATERIALS_DIR}/zookeeper/conf/log4j.properties
}


## FUNCS (BOOT)
function RUN_COMMAND() {
    command=$@
    INFO "Starting ${command}"
    sh -c "${command}" 2>&1
}

function INSTALL_ZK_CONF() {
    mkdir -p ${EQ_WORKING_DIR}/zk_conf ${EQ_WORKING_DIR}/zk_data
    chmod 777 ${EQ_WORKING_DIR}/zk_conf ${EQ_WORKING_DIR}/zk_data # FIXME
    cat > ${EQ_WORKING_DIR}/zk_conf/zoo.cfg <<EOF
tickTime=2000
initLimit=10
syncLimit=5
dataDir=${EQ_WORKING_DIR}/zk_data
clientPort=2181
EOF
}

function START_ZK() {
    RUN_COMMAND ${EQ_MATERIALS_DIR}/zookeeper/bin/zkServer-initialize.sh --myid=1 --configfile=${EQ_WORKING_DIR}/zk_conf/zoo.cfg
    RUN_COMMAND "${EQ_MATERIALS_DIR}/zookeeper/bin/zkServer.sh --config ${EQ_WORKING_DIR}/zk_conf start"
}

function BUILD_ZK_TEST() {
    ( cd ${EQ_MATERIALS_DIR};
      javac -cp $(find zookeeper/build -name '*.jar' | perl -pe 's/\n/:/g') MyZkCli/*.java
    )
}

function START_ZK_TEST() {
    ( cd ${EQ_MATERIALS_DIR};
      chmod -R 777 ${EQ_WORKING_DIR} # FIXME
      if [ -z $EQ_DISABLE ]; then
	  agent_jar="${EQ_MATERIALS_DIR}/earthquake-inspector.jar"
	  agent="-javaagent:${agent_jar}=script:${EQ_MATERIALS_DIR}/client.btm"
      else
	  agent=""
      fi
      cp="-cp MyZkCli:$(find zookeeper/build -name '*.jar' | perl -pe 's/\n/:/g') MyZkCli"
      RUN_COMMAND "EQ_ENV_ENTITY_ID=cli EQ_MODE_DIRECT=1 EQ_NO_INITIATION=1 java ${agent} ${cp} localhost:2181 2>&1 | tee ${EQ_WORKING_DIR}/zk-test.log"
    )
}

## FUNCS (SHUTDOWN)
function STOP_ZK() {
    RUN_COMMAND ${EQ_MATERIALS_DIR}/zookeeper/bin/zkServer.sh --config ${EQ_WORKING_DIR}/zk_conf stop
}

#!/bin/bash
set -x

EQDIR=$(pwd)/../..
ZKDIR=$(pwd)/zookeeper

export ZOOBINDIR=$ZKDIR/bin/
. $ZKDIR/bin/zkEnv.sh

AGENT_CP=$EQDIR/bin/earthquake-inspector.jar

for i in `seq 1 3`;
do
    DIR=/tmp/eq/zk/$i/
    if [ ! -d $DIR ];
    then
	mkdir -p $DIR
	echo $i > $DIR/myid
    fi

    export EQ_ENV_PROCESS_ID=zksrv$i 
    export SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:zk.btm" 
    export ZOO_LOG_DIR=/tmp/eq/zk-log/$i/ 
    mkdir -p $ZOO_LOG_DIR
    
    cp -r zk-conf /tmp/eq
    $ZKDIR/bin/zkServer.sh --config /tmp/eq/zk-conf/$i stop
done


rm -rf /tmp/eq/zk*   #NOTE: you must keep /tmp/eq/search

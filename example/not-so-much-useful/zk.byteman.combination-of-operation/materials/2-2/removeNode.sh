#! /bin/bash
set -x

export ZOOBINDIR=$EQ_MATERIALS_DIR/zookeeper/bin
. $ZOOBINDIR/zkEnv.sh

for i in `seq 1 2`;
do
    for t in `seq 1 5`;
    do
        echo "Remove Node"
        EQ_MODE_DIRECT=1 EQ_ENV_ENTITY_ID=zksrv$t EQ_NO_INITIATION=1 SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:$EQ_MATERIALS_DIR/server.btm" ZOO_LOG_DIR=$DIR/logs/$t/
        P=$((2180 + $t))
        java -cp $CLASSPATH:$EQ_MATERIALS_DIR/ReconfigZkCli ReconfigZkCli localhost $P
        if [ $? -eq 10 ]
        then
            break
        fi
    done
    sleep 5
done


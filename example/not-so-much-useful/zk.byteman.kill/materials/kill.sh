#! /bin/bash
echo ========== kill.sh enter ==========

if [ $# != 1 ];
then
    false
    return
fi

NID=""

if [ $1 == "zksrv1" ];
then
    NID="1"
fi

if [ $1 == "zksrv2" ];
then
    NID="2"
fi

if [ $1 == "zksrv3" ];
then
    NID="3"
fi

if [ $NID == "" ];
then
    false
    return
fi

kill -9 `cat $EQ_WORKING_DIR/zookeeper$NID/zookeeper_server.pid`

# shutdown
# $EQ_MATERIALS_DIR/bin/zkServer.sh --config $EQ_WORKING_DIR/quorumconf/$NID stop

# rebirth
# EQ_MODE_DIRECT=1 EQ_ENV_ENTITY_ID=zksrv$NID EQ_NO_INITIATION=1 SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:$EQ_MATERIALS_DIR/server.btm" ZOO_LOG_DIR=$DIR/logs/$NID/ $EQ_MATERIALS_DIR/bin/zkServer.sh --config $EQ_WORKING_DIR/quorumconf/$NID start

#! /bin/bash

export ZOOBINDIR=$EQ_MATERIALS_DIR/zookeeper/bin
. $ZOOBINDIR/zkEnv.sh

EQ_ENV_ENTITY_ID=zkcli1 EQ_MODE_DIRECT=1 EQ_NO_INITIATION=1 java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli -javaagent:$AGENT_CP=script:$EQ_MATERIALS_DIR/client.btm CreateZnodeZkCli localhost:2181 &

EQ_ENV_ENTITY_ID=zkcli2 EQ_MODE_DIRECT=1 EQ_NO_INITIATION=1 java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli -javaagent:$AGENT_CP=script:$EQ_MATERIALS_DIR/client.btm CreateZnodeZkCli localhost:2182


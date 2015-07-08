#! /bin/bash

export ZOOBINDIR=$MATERIALS_DIR/bin/
. $ZOOBINDIR/zkEnv.sh

EQ_ENV_PROCESS_ID=zkcli1 EQ_MODE_DIRECT=1 EQ_NO_INITIATION=1 java -cp $CLASSPATH:$MATERIALS_DIR/out/production/myZkCli -javaagent:$AGENT_CP=script:$MATERIALS_DIR/client.btm  MyZkCli localhost:2181 &

EQ_ENV_PROCESS_ID=zkcli2 EQ_MODE_DIRECT=1 EQ_NO_INITIATION=1 java -cp $CLASSPATH:$MATERIALS_DIR/out/production/myZkCli -javaagent:$AGENT_CP=script:$MATERIALS_DIR/client.btm  MyZkCli localhost:2182


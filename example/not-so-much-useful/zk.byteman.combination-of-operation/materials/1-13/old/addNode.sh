#! /bin/bash
set -x

export ZOOBINDIR=$EQ_MATERIALS_DIR/zookeeper/bin
. $ZOOBINDIR/zkEnv.sh

for i in `seq 1 2`;
do
    sleep 1
    P1=$((2183 + $i))
    P2=$((2890 + $i))
    P3=$((3890 + $i))
    NO=$((3 + $i))

    EQ_ENV_ENTITY_ID=zkcli3 EQ_MODE_DIRECT=1 EQ_NO_INITIATION=1 java -cp $CLASSPATH:$EQ_MATERIALS_DIR/AddNodeZkCli -javaagent:$AGENT_CP=script:$EQ_MATERIALS_DIR/client.btm AddNodeZkCli localhost:2181 server.$NO=localhost:$P2:$P3:participant\;$P1

done


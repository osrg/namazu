#! /bin/bash
set -x

EQDIR=$(pwd)/../..
ZKDIR=$(pwd)/zookeeper

export ZOOBINDIR=$ZKDIR/bin/
. $ZKDIR/bin/zkEnv.sh

AGENT_CP=$EQDIR/bin/earthquake-inspector.jar

for i in `seq 1 2`;
do
    export EQ_ENV_PROCESS_ID=zkcli$i 

    AGENT="-javaagent:$AGENT_CP=script:zk.btm" 
    jars=$(find . -name '*.jar' | perl -pe 's/\n/:/g')
    P=$((2180 + $i))
    java $AGENT -cp MyZkCli:$jars MyZkCli localhost:$P &
done

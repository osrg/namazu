#! /bin/bash
set -x

EQDIR=$(pwd)/../..
ZKDIR=$(pwd)/zookeeper

export ZOOBINDIR=$ZKDIR/bin/
. $ZKDIR/bin/zkEnv.sh

AGENT_CP=$EQDIR/bin/earthquake-inspector.jar

# createZnode
for i in `seq 1 2`;
do
    export EQ_ENV_PROCESS_ID=zkcli$i
    AGENT="-javaagent:$AGENT_CP=script:zk.btm"
    jars=$(find . -name '*.jar' | perl -pe 's/\n/:/g')
    P=$((2180 + $i))
    java $AGENT -cp MyZkCli:$jars MyZkCli localhost:$P &
done

# removeLeader & addRemovedLeader
export EQ_ENV_PROCESS_ID=zkcli3
AGENT="-javaagent:$AGENT_CP=script:zk.btm"
jars=$(find . -name '*.jar' | perl -pe 's/\n/:/g')

i=1
while :
do
    sleep 3
    P=$((2180 + $i))
    java $AGENT -cp ReconfigZkCli:$jars ReconfigZkCli localhost $P &
    i=$((i + 1))
    if [ ${i} -eq 4 ]
    then
        i=1
    fi
    if [ `ps -ef | grep MyZkCli | grep -v grep | wc -l` -eq 0 ]
    then
        break
    fi
done


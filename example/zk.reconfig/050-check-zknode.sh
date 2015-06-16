#! /bin/bash
set -x

EQDIR=$(pwd)/../..
ZKDIR=$(pwd)/zookeeper

export ZOOBINDIR=$ZKDIR/bin/
. $ZKDIR/bin/zkEnv.sh

for i in `seq 1 3`;
do
    DIR=/tmp/eq/zk/$i/
    if [ ! -d $DIR ];
    then
	mkdir -p $DIR
	echo $i > $DIR/myid
    fi

    export ZOO_LOG_DIR=/tmp/eq/zk-log/$i/ 
    mkdir -p $ZOO_LOG_DIR
    
    cp -r zk-conf /tmp/eq
    $ZKDIR/bin/zkServer.sh --config /tmp/eq/zk-conf/$i start
done

# checkZnode
sleep 5
jars=$(find . -name '*.jar' | perl -pe 's/\n/:/g')
java $AGENT -cp CheckZnodeZkCli:$jars CheckZnodeZkCli localhost:2181,localhost:2182,localhost:2183

for i in `seq 1 3`;
do
    $ZKDIR/bin/zkServer.sh --config /tmp/eq/zk-conf/$i stop
done

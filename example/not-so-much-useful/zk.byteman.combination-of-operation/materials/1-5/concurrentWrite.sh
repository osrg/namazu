#! /bin/bash
set -x

export ZOOBINDIR=$EQ_MATERIALS_DIR/zookeeper/bin
. $ZOOBINDIR/zkEnv.sh

java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183 &
java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183 &
java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183 &


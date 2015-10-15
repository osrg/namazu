#! /bin/bash
set -x

export ZOOBINDIR=$EQ_MATERIALS_DIR/zookeeper/bin
. $ZOOBINDIR/zkEnv.sh

#java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183,localhost:2184,localhost:2185 &
#java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183,localhost:2184,localhost:2185 &
#java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183,localhost:2184,localhost:2185 &
#java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183,localhost:2184,localhost:2185 &
#java -cp $CLASSPATH:$EQ_MATERIALS_DIR/CreateZnodeZkCli CreateZnodeZkCli localhost:2181,localhost:2182,localhost:2183,localhost:2184,localhost:2185 &

for i in `seq 1 10`;
do
    $EQ_MATERIALS_DIR/zookeeper/bin/zkCli.sh -server localhost:2181,localhost:2182,localhost:2183 create /test${i} bar
done


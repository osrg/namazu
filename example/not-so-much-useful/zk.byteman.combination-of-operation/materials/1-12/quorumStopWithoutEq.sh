#! /bin/bash

for i in `seq 1 1`;
do
    $EQ_MATERIALS_DIR/zookeeper/bin/zkServer.sh --config $EQ_WORKING_DIR/quorumconf/$i stop
done

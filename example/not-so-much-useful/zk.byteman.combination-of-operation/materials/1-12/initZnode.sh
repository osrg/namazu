#! /bin/bash
set -x

for i in `seq 1 10`;
do
    $EQ_MATERIALS_DIR/zookeeper/bin/zkCli.sh -server localhost:2181 create /test${i} bar
done



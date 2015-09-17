#! /bin/bash

for i in `seq 1 3`;
do
    echo Stopping zksrv$i
    $EQ_MATERIALS_DIR/bin/zkServer.sh --config $EQ_WORKING_DIR/quorumconf/$i stop
    echo Stopping zksrv$i
done

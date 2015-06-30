#! /bin/bash

for i in `seq 1 3`;
do
    $MATERIALS_DIR/bin/zkServer.sh --config $WORKING_DIR/quorumconf/$i stop
done

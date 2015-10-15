#! /bin/bash
set -x

bash $EQ_MATERIALS_DIR/quorumStop.sh
pkill -9 java

#sleep 1

#for i in `seq 1 5`;
#do
#    rm -rf $EQ_WORKING_DIR/zookeeper$i/version-2/*
#    rm -rf $EQ_WORKING_DIR/quorumconf/$i/zoo.cfg.dynamic.*
#    cp -R $EQ_WORKING_DIR/quorumconf/$i/zoo.cfg.bak $EQ_WORKING_DIR/quorumconf/$i/zoo.cfg
#    rm -rf $EQ_WORKING_DIR/quorumconf/$i/zoo.cfg.bak
#done



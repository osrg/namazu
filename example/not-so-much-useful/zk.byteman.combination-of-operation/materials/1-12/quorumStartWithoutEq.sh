#! /bin/bash

for i in `seq 1 1`;
do
    DIR=$EQ_WORKING_DIR/zookeeper$i/
    if [ ! -d $DIR ];
    then
	mkdir $DIR
	echo $i > $DIR/myid
    fi

    CFG=$EQ_WORKING_DIR/quorumconf/$i/zoo.cfg
    TMPCFG=$CFG.tmp
    echo "CFG:" $CFG
    echo "TMPCFG:" $TMPCFG
    sed "s#dataDir=#dataDir=$DIR#" $CFG > $TMPCFG
    mv $TMPCFG $CFG

    ZOO_LOG_DIR=$DIR/logs/$i/ $EQ_MATERIALS_DIR/zookeeper/bin/zkServer.sh --config $EQ_WORKING_DIR/quorumconf/$i start
done


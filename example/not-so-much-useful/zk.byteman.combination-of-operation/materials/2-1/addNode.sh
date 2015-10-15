#! /bin/bash
set -x

export ZOOBINDIR=$EQ_MATERIALS_DIR/zookeeper/bin
. $ZOOBINDIR/zkEnv.sh

for i in `seq 4 5`;
do
    DIR=$EQ_WORKING_DIR/zookeeper$i/
    if [ ! -d $DIR ];
    then
        mkdir $DIR
        echo $i > $DIR/myid

        CFG=$EQ_WORKING_DIR/quorumconf/$i/zoo.cfg
        TMPCFG=$CFG.tmp
        echo "CFG:" $CFG
        echo "TMPCFG:" $TMPCFG
        sed "s#dataDir=#dataDir=$DIR#" $CFG > $TMPCFG
        mv $TMPCFG $CFG
    fi

    EQ_MODE_DIRECT=1 EQ_ENV_ENTITY_ID=zksrv$i EQ_NO_INITIATION=1 SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:$EQ_MATERIALS_DIR/server.btm" ZOO_LOG_DIR=$DIR/logs/$i/ $EQ_MATERIALS_DIR/zookeeper/bin/zkServer.sh --config $EQ_WORKING_DIR/quorumconf/$i start

    sleep 1
    P1=$((2180 + $i))
    P2=$((2887 + $i))
    P3=$((3887 + $i))
    NO=$(($i))

    java -cp $CLASSPATH:$EQ_MATERIALS_DIR/AddNodeZkCli AddNodeZkCli localhost:2181 server.$NO=localhost:$P2:$P3:participant\;$P1
done


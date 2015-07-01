#! /bin/bash

function RECONFIG_ADD_SERVER(){
    sid=$1                                                                                                                           
    trials=$2
    sleep_secs=$3
    for f in $(seq 1 ${trials}); do
        INFO "Reconfig (sid=${sid}, trial=${f} of ${trials})"
        tmp=$(mktemp)
        cmd="$MATERIALS_DIR/bin/zkCli.sh -server localhost:2181 reconfig -add server.${sid}=localhost:$((2887+sid)):$((3887+sid)):participant;$((2180+sid))"
        INFO "Reconfig Invoking: ${cmd}"
        ${cmd} 2>&1 | tee ${tmp}
        errors=$(grep KeeperErrorCode ${tmp} | wc -l)
        if [ $errors -eq 0 ]; then
            INFO "Reconfig success (sid=${sid})"; rm -f ${tmp}; return $(true)
        fi
        INFO "Reconfig fail (sid=${sid}, trial=${f} of ${trials})"
        rm -f $tmp
        SLEEP ${sleep_secs}
    done
    return $(false)
}

for i in `seq 1 3`;
do
    DIR=$WORKING_DIR/zookeeper$i/
    if [ ! -d $DIR ];
    then
	mkdir $DIR
	echo $i > $DIR/myid
    fi

    CFG=$WORKING_DIR/quorumconf/$i/zoo.cfg
    TMPCFG=$CFG.tmp
    echo "CFG:" $CFG
    echo "TMPCFG:" $TMPCFG
    sed "s#dataDir=#dataDir=$DIR#" $CFG > $TMPCFG
    mv $TMPCFG $CFG
done

EQ_MODE_DIRECT=1 EQ_ENV_PROCESS_ID=zksrv1 EQ_NO_INITIATION=1 SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:$MATERIALS_DIR/server_slowdb.btm" ZOO_LOG_DIR=$WORKING_DIR/zookeeper1/logs/1/ $MATERIALS_DIR/bin/zkServer.sh --config $WORKING_DIR/quorumconf/1 start

EQ_MODE_DIRECT=1 EQ_ENV_PROCESS_ID=zksrv2 EQ_NO_INITIATION=1 SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:$MATERIALS_DIR/server_slowdb.btm" ZOO_LOG_DIR=$WORKING_DIR/zookeeper2/logs/2/ $MATERIALS_DIR/bin/zkServer.sh --config $WORKING_DIR/quorumconf/2 start

RECONFIG_ADD_SERVER 2 10 5

EQ_MODE_DIRECT=1 EQ_ENV_PROCESS_ID=zksrv3 EQ_NO_INITIATION=1 SERVER_JVMFLAGS="-javaagent:$AGENT_CP=script:$MATERIALS_DIR/server_slowdb.btm" ZOO_LOG_DIR=$WORKING_DIR/zookeeper3/logs/3/ $MATERIALS_DIR/bin/zkServer.sh --config $WORKING_DIR/quorumconf/3 start

RECONFIG_ADD_SERVER 3 5 5

true


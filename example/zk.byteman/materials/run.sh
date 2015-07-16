#! /bin/bash

export ZOOBINDIR=$EQ_MATERIALS_DIR/bin/
. $ZOOBINDIR/zkEnv.sh

export AGENT_CP=$EQ_MATERIALS_DIR/earthquake-inspector.jar

cp -R $EQ_MATERIALS_DIR/quorumconf.template $EQ_WORKING_DIR/quorumconf

bash $EQ_MATERIALS_DIR/quorumStart.sh
bash $EQ_MATERIALS_DIR/concurrentWrite.sh
sleep 1
bash $EQ_MATERIALS_DIR/quorumStop.sh


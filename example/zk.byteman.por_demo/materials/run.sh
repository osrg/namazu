#! /bin/bash

export ZOOBINDIR=$MATERIALS_DIR/bin/
. $ZOOBINDIR/zkEnv.sh

export AGENT_CP=$MATERIALS_DIR/earthquake-inspector.jar

cp -R $MATERIALS_DIR/quorumconf.template $WORKING_DIR/quorumconf

bash $MATERIALS_DIR/quorumStart.sh
bash $MATERIALS_DIR/concurrentWrite.sh

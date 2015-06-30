#! /bin/bash

export ZOOBINDIR=$MATERIALS_DIR/bin/
. $ZOOBINDIR/zkEnv.sh

export AGENT_CP=$MATERIALS_DIR/earthquake-inspector.jar

cp -R $MATERIALS_DIR/2172_confs $WORKING_DIR/quorumconf

bash $MATERIALS_DIR/2172_start.sh

#! /bin/bash

export ZOOBINDIR=$EQ_MATERIALS_DIR/bin/
. $ZOOBINDIR/zkEnv.sh

export AGENT_CP=$EQ_MATERIALS_DIR/earthquake-inspector.jar

cp -R $EQ_MATERIALS_DIR/2172_confs $EQ_WORKING_DIR/quorumconf

bash $EQ_MATERIALS_DIR/2172_start.sh

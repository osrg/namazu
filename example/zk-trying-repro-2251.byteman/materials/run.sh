#!/bin/bash
set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

## some bad hack: allows modifying materials for interactive experiments
cp ${EQ_MATERIALS_DIR}/../config.json ${EQ_WORKING_DIR}/config.json.BAK
cp -r ${EQ_MATERIALS_DIR}/MyZkCli ${EQ_WORKING_DIR}/MyZkCli.BAK

########## Boot ##########
# build here rather than init for interactive experiments
BUILD_ZK_TEST

INSTALL_ZK_CONF
START_ZK
START_ZK_TEST || true
#PAUSE

########## Shutdown ##########
STOP_ZK

exit 0

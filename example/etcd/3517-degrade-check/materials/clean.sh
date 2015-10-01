#!/bin/bash
#set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

# CLEAN_VETHS # old pipework needs CLEAN_VETHS

########## Shutdown ##########
KILL_DOCKER
if [ -z $EQ_DISABLE ]; then
    KILL_SWITCH
    KILL_INSPECTOR
fi

INFO "Please run \"docker rmi ${DOCKER_IMAGE_NAME}\" if needed"

exit 0

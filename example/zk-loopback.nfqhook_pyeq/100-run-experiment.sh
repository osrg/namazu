#!/bin/bash
set -e # exit on an error
source lib.sh

EXP_DIR=/tmp/eq-run-experiment/$(date +"%Y%m%d.%H%M%S")
mkdir -p ${EXP_DIR}
IMPORTANT "===== EXPERIMENT BEGIN (${EXP_DIR}) ====="

DO_TEST || (IMPORTANT "PERHAPS THE BUG WAS REPRODUCED!"; touch ${EXP_DIR}/REPRODUCED)
mv ${TEST_LOGFILE} ${EXP_DIR}/log

if [ -z $DISABLE_EQ ]; then
    INFO "Stopping inspection"
    STOP_EQ_INSPECTION
    last_exp=$(ls /tmp/eq/search/history | tail -1)
    cp -r /tmp/eq/search/history/${last_exp} ${EXP_DIR}
fi

if [ -e ${EXP_DIR}/REPRODUCED ]; then
    IMPORTANT "PLEASE CHECK WHETHER THE BUG WAS REPRODUCED!!"
fi    


IMPORTANT "===== EXPERIMENT END (${EXP_DIR}) ====="

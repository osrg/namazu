#!/bin/bash
#set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

result=$(cat ${EQ_WORKING_DIR}/check-fle-states.result)
INFO "result: ${result}"
exit ${result}

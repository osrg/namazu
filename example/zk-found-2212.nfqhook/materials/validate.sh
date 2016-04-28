#!/bin/bash
#set -e # exit on an error
. ${NMZ_MATERIALS_DIR}/lib.sh

result=$(cat ${NMZ_WORKING_DIR}/check-fle-states.result)
INFO "result: ${result}"
exit ${result}

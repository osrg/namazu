#!/bin/bash
#set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

# TODO: use etcdctl
echo "creating key"
etcdctl -C http://192.168.42.1:4001 set /k v
echo "result: $?"

# result=$(cat ${EQ_WORKING_DIR}/check-fle-states.result)
# INFO "result: ${result}"
# exit ${result}

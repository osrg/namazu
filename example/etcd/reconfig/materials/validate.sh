#!/bin/bash
#set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

export ETCDCTL_PEERS=http://192.168.42.1:4001,http://192.168.42.2:4001,http://192.168.42.3:4001
echo "creating key"
etcdctl --no-sync set /k v
result=$?
echo "result: $result"

exit $result

# result=$(cat ${EQ_WORKING_DIR}/check-fle-states.result)
# INFO "result: ${result}"
# exit ${result}

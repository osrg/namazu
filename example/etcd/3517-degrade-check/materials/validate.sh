#!/bin/bash
. ${NMZ_MATERIALS_DIR}/lib.sh

export ETCDCTL_PEERS=http://192.168.42.1:4001,http://192.168.42.2:4001,http://192.168.42.3:4001
echo "getting key"
$NMZ_MATERIALS_DIR/etcd_testbed/etcd/bin/etcdctl --no-sync --timeout 10s --total-timeout 10s get /k1
result=$?
echo "result: $result"

exit $result

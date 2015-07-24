#! /bin/bash
set -x

EQDIR=$(pwd)/../..
ZKDIR=$(pwd)/zookeeper

export ZOOBINDIR=$ZKDIR/bin/
. $ZKDIR/bin/zkEnv.sh
export EQ_ENV_PROCESS_ID=

jars=$(find . -name '*.jar' | perl -pe 's/\n/:/g')
java $AGENT -cp CheckZnodeZkCli:$jars CheckZnodeZkCli localhost:2181,localhost:2182,localhost:2183


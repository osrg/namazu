#! /bin/bash

export ZOOBINDIR=/home/mitake/github/zookeeper.git/bin/
. $ZOOBINDIR/zkEnv.sh

java -cp $CLASSPATH:./MyZkCli/out/production/MyZkCli MyZkCli localhost:2181 localhost:2182 localhost:2183

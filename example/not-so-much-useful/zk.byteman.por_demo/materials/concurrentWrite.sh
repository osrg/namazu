#! /bin/bash

. $ZOOBINDIR/zkEnv.sh
java -cp $CLASSPATH:$EQ_MATERIALS_DIR/MyZkCli/out/production/MyZkCli MyZkCli localhost:2181 localhost:2182 localhost:2183


#!/bin/bash
set -x

# download zookeeper to this dir
git clone https://github.com/apache/zookeeper.git
pushd zookeeper
ant
popd

javac -cp $(find . -name '*.jar' | perl -pe 's/\n/:/g') MyZkCli/*.java
javac -cp $(find . -name '*.jar' | perl -pe 's/\n/:/g') ReconfigZkCli/*.java   
javac -cp $(find . -name '*.jar' | perl -pe 's/\n/:/g') CheckZnodeZkCli/*.java

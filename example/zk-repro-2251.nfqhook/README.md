# ZooKeeper Bug [ZOOKEEPER-2251](https://issues.apache.org/jira/browse/ZOOKEEPER-2251): Client gets hanged


## How to Reproduce the Bug with Earthquake

This directory does not include a reproduction script for ZOOKEEPER-2251 yet.

    $ docker run -it ubuntu:14.04
    docker$ apt-get update && apt-get install zookeeper # tested with 3.4.5
	docker$ PATH=/usr/share/zookeeper/bin:$PATH
	docker$ zkServer.sh start
	docker$ SETUP_EARTHQUAKE # to be documented. Dumb exploration policy is enough.
    docker$ for f in $(seq 1 1000); do zkCli.sh ls /; done


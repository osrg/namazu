# ZooKeeper Bug [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080): ReconfigRecoveryTest fails intermittently

The JUnit test `ReconfigRecoveryTest.testCurrentObserverIsParticipantInNewConfig` fails intermittently.

The cause of the bug had been unknown for almost 2 years (since 2013).

Using Namazu, we reproduced the bug and analyzed its cause.

The bug has been caused by race conditions in `QuorumCnxManager`, which manages sockets for leader election.

To fix the bug gracefully, drastic changes to `QuorumCnxManager` (suggested since 2010: [ZOOKEEPER-901](https://issues.apache.org/jira/browse/ZOOKEEPER-901)) might be needed. 

## How to Reproduce the Bug with Namazu

The bug can be easily reproduced by injecting several tens of millisecs sleeps to FLE packets.

### Start Namazu
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.


	$ sudo sh -c 'echo 0 > /proc/sys/net/ipv4/tcp_autocorking' # recommended if you are using Linux 3.14 or later
	$ sudo useradd -m nfqhooked # this user is needed for internal sandboxing
	$ sudo PYTHONPATH=$(pwd)/../../misc ../../bin/nmz init --force config.toml materials /tmp/zk-2080

Disabling `tcp_autocorking` is recommended for future zktraffic-based inspection support.

Currently zktraffic does not work well with ZOOKEEPER-2080 scenario, because some packets still somehow get corked even when `tcp_autocorking` is disabled.

### Run Experiments

	$ sudo PYTHONPATH=$(pwd)/../../misc ../../bin/nmz run /tmp/zk-2080

If the bug could not be reproduced, you might have to modify the `interval` parameter in `config.toml`. (about 30 msecs to 80 msecs)

## Analyze
Unlike [ZOOKEEPER-2212](../zk-found-2212.ryu/README.md), you cannot use the Namazu event history for analysis.

Instead, you can estimate the cause of the bug using Namazu branch analyzer for Java.

	
	$ java -jar ../../bin/nmz-analyzer.jar /tmp/zk-2080/ --classes-path /tmp/zk-2080/materials/zookeeper/build/classes
	[DEBUG] net.osrg.namazu.ExperimentAnalyzer - Scanning /tmp/zk-2080/00000000: experiment successful=false
	[DEBUG] net.osrg.namazu.ExperimentAnalyzer - Scanning /tmp/zk-2080/00000001: experiment successful=true
	..
	Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::connectAll
		- at line 511: branch on success=0, on failure=4
	..
	Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::receiveConnection
		- at line 358: branch on success=0, on failure=4
	..
	

### Environment Variables

* `NMZ_DISABLE`(default: (unset)): disable the substantial part of Namazu if set
* `ZK_GIT_COMMIT`(default:(see `materials/lib.sh`)) : use another ZooKeeper version
* `ZK_SOURCE_DIR`(default: (unset)) : use another ZooKeeper source directory if set

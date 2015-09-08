# ZooKeeper Bug [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080): ReconfigRecoveryTest fails intermittently

The JUnit test `ReconfigRecoveryTest.testCurrentObserverIsParticipantInNewConfig` fails intermittently.

The cause of the bug had been unknown for almost 2 years (since 2013).

Using Earthquake, we reproduced the bug and analyzed its cause.

The bug has been caused by race conditions in `QuorumCnxManager`, which manages sockets for leader election.

To fix the bug gracefully, drastic changes to `QuorumCnxManager` (suggested since 2010: [ZOOKEEPER-901](https://issues.apache.org/jira/browse/ZOOKEEPER-901)) might be needed. 

## How to Reproduce the Bug with Earthquake

The bug can be easily reproduced by injecting several tens of millisecs sleeps to FLE packets.

### Start Earthquake
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.


	$ sudo sh -c 'echo 0 > /proc/sys/net/ipv4/tcp_autocorking' # recommended if you are using Linux 3.14 or later
	$ sudo useradd -m nfqhooked # this user is needed for internal sandboxing
	$ sudo PYTHONPATH=$(pwd)/../.. ../../bin/earthquake init --force config.toml materials /tmp/zk-2080

Disabling `tcp_autocorking` is recommended for future zktraffic-based inspection support.

Currently zktraffic does not work well with ZOOKEEPER-2080 scenario, because some packets still somehow get corked even when `tcp_autocorking` is disabled.

### Run Experiments

	$ sudo PYTHONPATH=$(pwd)/../.. ../../bin/earthquake run /tmp/zk-2080

If the bug could not be reproduced, you might have to modify the `sleep` parameter in `config.toml`. (about 30 msecs to 80 msecs)

## Analyze
TBD

### Environment Variables

* `EQ_DISABLE`(default: (unset)): disable the substantial part of Earthquake if set
* `ZK_GIT_COMMIT`(default:(see `materials/lib.sh`)) : use another ZooKeeper version
* `ZK_SOURCE_DIR`(default: (unset)) : use another ZooKeeper source directory if set

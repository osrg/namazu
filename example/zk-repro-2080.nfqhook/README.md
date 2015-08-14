# ZooKeeper Bug [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080): ReconfigRecoveryTest fails intermittently


## How to Reproduce the Bug with Earthquake

The bug can be easily reproduced by injecting several tens of millisecs sleeps to FLE packets.

### Start Earthquake
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.


	$ sudo sh -c 'echo 0 > /proc/sys/net/ipv4/tcp_autocorking' # recommended if you are using Linux 3.14 or later
	$ sudo useradd --foo-bar nfqhooked
	$ sudo iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num 42
	$ sudo PYTHONPATH=$(pwd)/../.. ../../bin/earthquake init --force config.toml materials /tmp/zk-2080

Disabling `tcp_autocorking` is recommended for future zktraffic-based inspection support.

Currently zktraffic does not work well with ZOOKEEPER-2080 scenario, because some packets still somehow get corked even when `tcp_autocorking` is disabled.

### Run Experiments

	$ sudo PYTHONPATH=$(pwd)/../.. ../../bin/earthquake run /tmp/zk-2080

If the bug could not be reproduced, you might have to modify the `sleep` parameter in `config.toml`. (about 30 msecs to 80 msecs)


### Environment Variables

* `EQ_DISABLE`(default: (unset)): disable the substantial part of Earthquake if set
* `ZK_GIT_COMMIT`(default:(see `materials/lib.sh`)) : use another ZooKeeper version
* `ZK_SOURCE_DIR`(default: (unset)) : use another ZooKeeper source directory if set
* `SYSLOG_DISABLE`(default: (unset)): disable Syslog inspector if set


# ZooKeeper Bug [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080): ReconfigRecoveryTest fails intermittently


## How to Reproduce the Bug with Earthquake
   
NOTE: this test scenario is under work-in-progress. reproduction probability is still not so high.
   
### Start Earthquake
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.


	$ sudo pip install pip install git+https://github.com/twitter/zktraffic@68d9f85d8508e01f5d2f6657666c04e444e6423c  #(Jul 18, 2015)
	$ sudo sh -c 'echo 0 > /proc/sys/net/ipv4/tcp_autocorking' # required if you are using Linux 3.14 or later
	$ sudo useradd --foo-bar nfqhooked
	$ sudo iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num 42
    $ sudo PYTHONPATH=$(pwd)/../.. ../../bin/earthquake init --force config.toml materials /tmp/zk-2080


### Run Experiments
    
    $ sudo ../../bin/earthquake run /tmp/zk-2080

### Environment Variables

 * `EQ_DISABLE`: disable the substantial part of Earthquake if set
 * `ZK_GIT_COMMIT`: use another ZooKeeper version


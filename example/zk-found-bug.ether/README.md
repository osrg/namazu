# ZooKeeper Bug: distributed race condition related to QV version

When a joiner is listed as an observer in an initial config,
the joiner should become a non-voting follower (not an observer) until reconfig is triggered. [(Link)](http://zookeeper.apache.org/doc/trunk/zookeeperReconfig.html#sc_reconfig_general)

I found a race-condition situation where an observer keeps being an observer and cannot become a non-voting follower.

This race condition happens when an observer receives an UPTODATE Quorum Packet from the leader:2888/tcp *after* receiving a Notification FLE Packet of which n.config version is larger than the observer's one from leader:3888/tcp.

Event history: [example-output/3.REPRODUCED/json](example-output/3.REPRODUCED/json)

## ZooKeeper Version
commit 98a3cabfa279833b81908d72f1c10ee9f598a045 (Tue Jun 2 19:17:09 2015 +0000)

## Detail
 * Problem: An observer cannot become a non-voting follower
 * Cause: Cannot restart FLE
 * Cause: In QuorumPeer.run(), cannot shutdown Observer [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/QuorumPeer.java#L1014)
 * Cause: In QuorumPeer.run(), cannot return from Observer.observeLeader()  [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/QuorumPeer.java#L1010)
 * Cause: In Observer.observeLeader(), Learner.syncWithLeader() does not throw an exception of "changes proposed in reconfig" [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/Observer.java#L79)
 * Cause: In Learner.syncWithLeader(), QuorumPeer.processReconfig() returns false with a log message like "2 setQuorumVerifier called with known or old config 4294967296. Current version: 4294967296".
 * Cause: The observer have already received a Notification Packet(n.config.version=4294967296), invoked QuorumPeer.processReconfig() [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/FastLeaderElection.java#L291-304)
   

## How to reproduce the bug with earthquake
    
### Start Earthquake
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.

    $ sudo pip install zktraffic==0.1.3
    $ ./000-prepare.sh
    [INFO] Starting Earthquake Ethernet Switch
    [INFO] Starting Earthquake Orchestrator
    [INFO] Starting Earthquake Ethernet Inspector
    [IMPORTANT] Please kill the processes (switch=1234, orchestrator=1235, and inspector=1236) after you finished all of the experiments
    [IMPORTANT] Please continue to 100-run-experiment.sh..
    

### Run Experiments
    
    $ ./100-run-experiment.sh
    $ ./100-run-experiment.sh
    .. #(try several times, about 5 times or more)
    $ ./100-run-experiment.sh
    [IMPORTANT] THE BUG WAS REPRODUCED!
    $ kill -9 1234 1235 1236
    

NOTE: You can set `export DISABLE_EQ=1` before `./100-prepare.sh` to disable earthquake.

### Change Zookeeper Version
    
    $ ( cd zk_testbed/zookeeper; git checkout FOOBAR)
    $ docker rmi zk_testbed
    $ ./000-prepare.sh
    

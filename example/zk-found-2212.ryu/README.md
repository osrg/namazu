# ZooKeeper Bug [ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212): distributed race condition related to QV version

When a joiner is listed as an observer in an initial config,
the joiner should become a non-voting follower (not an observer) until reconfig is triggered. [(Link)](http://zookeeper.apache.org/doc/trunk/zookeeperReconfig.html#sc_reconfig_general)

We found a race-condition situation where an observer keeps being an observer and cannot become a non-voting follower.

This race condition happens when an observer receives an `UPTODATE` Quorum Packet from the leader:2888/tcp *after* receiving a `Notification` FLE Packet of which n.config version is larger than the observer's one from leader:3888/tcp.

## ZooKeeper Version
[commit 98a3cabfa279833b81908d72f1c10ee9f598a045 (Tue Jun 2 19:17:09 2015 +0000)](https://github.com/apache/zookeeper/commit/98a3cabfa279833b81908d72f1c10ee9f598a045)

NOTE: We reported the bug ([ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212)) to ZooKeeper community, and the bug is fixed in [commit ec056d3c3a18b862d0cd83296b7d4319652b0b1c (Mon Jun 15 23:05:25 2015 +0000)](https://github.com/apache/zookeeper/commit/ec056d3c3a18b862d0cd83296b7d4319652b0b1c).

## Details
 * Problem: An observer cannot become a non-voting follower
 * Cause: Cannot restart FLE
 * Cause: In QuorumPeer.run(), cannot shutdown Observer [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/QuorumPeer.java#L1014)
 * Cause: In QuorumPeer.run(), cannot return from Observer.observeLeader()  [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/QuorumPeer.java#L1010)
 * Cause: In Observer.observeLeader(), Learner.syncWithLeader() does not throw an exception of "changes proposed in reconfig" [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/Observer.java#L79)
 * Cause: In Learner.syncWithLeader(), QuorumPeer.processReconfig() returns false with a log message like "2 setQuorumVerifier called with known or old config 4294967296. Current version: 4294967296".
 * Cause: The observer have already received a Notification Packet(n.config.version=4294967296), invoked QuorumPeer.processReconfig() [(Link)](https://github.com/apache/zookeeper/blob/98a3cabfa279833b81908d72f1c10ee9f598a045/src/java/main/org/apache/zookeeper/server/quorum/FastLeaderElection.java#L291-304)
   

## How to Reproduce the Bug with Namazu
    
### Start Namazu
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.

The pre-built Docker image (`osrg/namazu`) is strongly recommended, 
because `ovsbr0` is expected to be configured as `192.168.42.254/24` in the experiments.

NOTE: If git master version is corrupted, you can use [osrg/earthquake-zookeeper-2212](https://registry.hub.docker.com/u/osrg/earthquake-zookeeper-2212/) container (based on Namazu v0.1).

    $ sudo pip install git+https://github.com/twitter/zktraffic@68d9f85d8508e01f5d2f6657666c04e444e6423c  #(Jul 18, 2015)
    $ sudo pip install hookswitch
    $ sudo PYTHONPATH=$(pwd)/../../misc ../../bin/nmz init --force config.toml materials /tmp/zk-2212
    [INFO] Checking whether Docker is installed
    [INFO] Checking whether pipework is installed
    [INFO] Checking whether ryu is installed
    [INFO] Checking whether ovsbr0 is configured as 192.168.42.254
    [INFO] Fetching ZooKeeper
    [INFO] Checking out ZooKeeper@98a3cabfa279833b81908d72f1c10ee9f598a045
    [INFO] You can change the ZooKeeper version by setting ZK_GIT_COMMIT
    [INFO] Building Docker Image zk_testbed (/tmp/hoge/materials/docker-build.log)
    ok


### Run Experiments
    
    $ sudo PYTHONPATH=$(pwd)/../../misc ../../bin/nmz run /tmp/zk-2212
    [INFO] Checking PYTHONPATH(=/home/suda/WORK/namazu/example/zk-found-2212.ryu/../../misc)
    [INFO] Starting Namazu Ethernet Switch
    [INFO] Switch PID: 28893
    [INFO] Starting Namazu Ethernet Inspector
    [INFO] Inspector PID: 28894
    [INFO] Starting Docker container zk1 from zk_testbed
    [INFO] Starting Docker container zk2 from zk_testbed
    [INFO] Starting Docker container zk3 from zk_testbed
    [INFO] Assigning 192.168.42.1/24 (ovsbr0) to zk1
    [INFO] Assigning 192.168.42.2/24 (ovsbr0) to zk2
    [INFO] Assigning 192.168.42.3/24 (ovsbr0) to zk3
    [INFO] Starting ZooKeeper(sid=1) in Docker container zk1
    [INFO] Starting ZooKeeper(sid=2) in Docker container zk2
    [INFO] Starting ZooKeeper(sid=3) in Docker container zk3
    [INFO] Sleeping(5 secs)..
    [INFO] Checking FLE states
    [IMPORTANT] Failure: 1 (/tmp/zk-2212/00000002/check-fle-states.log) # this failure means that the bug is reproduced
    [INFO] Killing Docker container zk1 (log:/tmp/zk-2212/00000002/zk1)
    [INFO] Killing Docker container zk2 (log:/tmp/zk-2212/00000002/zk2)
    [INFO] Killing Docker container zk3 (log:/tmp/zk-2212/00000002/zk3)
    [INFO] Killing Switch, PID: 28893
    [INFO] Killing Inspector, PID: 28894
    [INFO] result: 1
    validation failed: exit status 1


    
You may have to run the experiments for 3 or 5 times.

You can check which experiment reproduced the bug as follows:

    $ sudo ../../bin/nmz tools summary /tmp/zk-2212
    Fri Jul 24 19:46:15 JST 2015 ...orage/naive/naive.go(142): a number of collected traces: 3
    00000002 caused failure

### Example Result

[example-result.20150805](example-result.20150805) is an example result of `/tmp/zk-2212` in the above scenario.

#### [00000000](example-result.20150805/00000000): *not* reproduced the bug
zk2 was successfully promoted to an observer to a follower, because it received `UpToDate` before `Notification(config.version=100000000)`

* [32](example-result.20150805/00000000/actions/32.event.json): zk1->zk2: `UpToDate`
* [36](example-result.20150805/00000000/actions/36.event.json): zk1<-zk2: `FollowerInfo`
* [37](example-result.20150805/00000000/actions/37.event.json): zk1<-zk2: `Notification(config.version=100000000)`
* [39](example-result.20150805/00000000/actions/39.event.json): zk1->zk2: `Notification(config.version=100000000)`


zk3 was successfully promoted to an observer to a follower, because it received `UpToDate` before `Notification(config.version=100000000)`

* [20](example-result.20150805/00000000/actions/20.event.json): zk1->zk3: `UpToDate`
* [21](example-result.20150805/00000000/actions/21.event.json): zk1<-zk3: `Notification(config.version=100000000)`
* [23](example-result.20150805/00000000/actions/23.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [25](example-result.20150805/00000000/actions/25.event.json): zk1<-zk3: `FollowerInfo`

#### [00000001](example-result.20150805/00000001): *not* reproduced the bug
zk2 was already a follower when it received `UpToDate`

* [19](example-result.20150805/00000001/actions/19.event.json): zk1<-zk2: `FollowerInfo`
* [31](example-result.20150805/00000001/actions/31.event.json): zk1->zk2: `UpToDate`

zk3 was successfully promoted to an observer to a follower, because it received `UpToDate` before `Notification(config.version=100000000)`

* [26](example-result.20150805/00000001/actions/26.event.json): zk1->zk3: `UpToDate`
* [29](example-result.20150805/00000001/actions/29.event.json): zk1<-zk3: `FollowerInfo`
* [34](example-result.20150805/00000001/actions/34.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [37](example-result.20150805/00000001/actions/37.event.json): zk1<-zk3: `Notification(config.version=100000000)`

#### [00000002](example-result.20150805/00000002): *reproduced* the bug (zk2, zk3)
zk2 was not able to be promoted (please see above for the reason)

* [12](example-result.20150805/00000002/actions/12.event.json): zk1->zk2: `Notification(config.version=100000000)`
* [19](example-result.20150805/00000002/actions/19.event.json): zk1->zk2: `UpToDate`
* No `FollowerInfo` (zk1<-zk2)

zk3 was not able to be promoted (please see above for the reason)

* [10](example-result.20150805/00000002/actions/10.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [24](example-result.20150805/00000002/actions/24.event.json): zk1->zk3: `UpToDate`
* No `FollowerInfo` (zk1<-zk3)

#### [00000003](example-result.20150805/00000003): *reproduced* the bug (zk3)
zk2 was successfully promoted to an observer to a follower, because it received `UpToDate` before `Notification(config.version=100000000)`

* [24](example-result.20150805/00000003/actions/24.event.json): zk1->zk2: `UpToDate`
* [25](example-result.20150805/00000003/actions/25.event.json): zk1<-zk2: `Notification(config.version=100000000)`
* [27](example-result.20150805/00000003/actions/27.event.json): zk1->zk2: `Notification(config.version=100000000)`
* [29](example-result.20150805/00000003/actions/29.event.json): zk1<-zk2: `FollowerInfo`

zk3 was not able to be promoted (please see above for the reason)

* [7](example-result.20150805/00000003/actions/7.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [20](example-result.20150805/00000003/actions/20.event.json): zk1->zk3: `UpToDate`
* No `FollowerInfo` (zk1<-zk3)

### config.toml
Experimental feature: You can also store the result in MongoDB by setting `storageType` to `mongodb` in `config.toml`.

### Environment Variables

 * `NMZ_DISABLE`(default:(unset)): disable the substantial part of Namazu if set. When Namazu is disabled, we could not reproduced the bug in 3 days.
 * `ZK_GIT_COMMIT`(default:98a3ca..): use another ZooKeeper version
 * `ZK_START_WAIT_SECS`(default:10): should be increased if there is false-positive
 * `PAUSE_ON_FAILURE`(default:0): pause on a possible failure for interactive verification if set to 1


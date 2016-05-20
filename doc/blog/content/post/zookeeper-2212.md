+++
categories = ["blog"]
date = "2015-08-20"
tags = ["document","found-bug","zookeeper"]
title = "ZOOKEEPER-2212: distributed race condition"

+++

## Introduction
[Apache Zookeeper](https://zookeeper.apache.org/) is a highly available coordination service that is used by distributed systems such as [Hadoop(via YARN)](https://hadoop.apache.org/), [Spark](http://spark.apache.org/),  [Mesos](http://mesos.apache.org/), [Kafka](http://kafka.apache.org/), [HBase](http://hbase.apache.org/), and many others.

Using Earthquake, we found a distributed race-condition bug of ZooKeeper, which can lead to service unavailability.

We reported the bug to ZooKeeper community ([ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212)), and the bug is fixed in [commit ec056d (Jun 15, 2015)](https://github.com/apache/zookeeper/commit/ec056d3c3a18b862d0cd83296b7d4319652b0b1c).


## The Bug
We found a race-condition situation where an "observer" server keeps being an observer and cannot become a "participant". [(Document)](http://zookeeper.apache.org/doc/trunk/zookeeperReconfig.html#sc_reconfig_general)

This race condition happens when an observer receives an `UPTODATE` ZAB (ZooKeeper Atomic Broadcast protocl) packet from the leader:2888/tcp *after* receiving a `Notification` FLE (Fast Leader Election protocol) packet of which n.config version is larger than the observer's one from leader:3888/tcp.

Without Earthquake, we could not reproduce the bug in 5,000 experiments. (took about 60 hours)


## How to Reproduce the Bug with Earthquake
    
### Set up Earthquake (v0.1.1)
Please see [doc/how-to-setup-env.md](https://github.com/osrg/namazu/blob/v0.1.1/doc/how-to-setup-env.md) for how to setup the environment.

The use of pre-built Docker image `osrg/earthquake:v0.1.1` is strongly recommended, which saves you the labor for setting up Open vSwitch and ryu.

    $ sudo modprobe openvswitch # tested with Ubuntu 15.04 host (Linux kernel 3.19)
    $ docker run --rm --tty --interactive --privileged -e EQ_DOCKER_PRIVILEGED=1 osrg/earthquake:v0.1.1


Then, build ZooKeeper "Docker-in-Docker" containers, and initialize Earthquake as follows.

    docker$ pip install git+https://github.com/twitter/zktraffic@68d9f85d8508e01f5d2f6657666c04e444e6423c  #(Jul 18, 2015)
    docker$ export PYTHONPATH=/earthquake
    docker$ cd /earthquake/example/zk-found-2212.ryu
    docker$ ../../bin/earthquake init --force config.toml materials /tmp/zk-2212
    [INFO] Checking whether Docker is installed
    [INFO] Checking whether pipework is installed
    [INFO] Checking whether ryu is installed
    [INFO] Checking whether ovsbr0 is configured as 192.168.42.254
    [INFO] Fetching ZooKeeper
    [INFO] Checking out ZooKeeper@98a3cabfa279833b81908d72f1c10ee9f598a045
    [INFO] You can change the ZooKeeper version by setting ZK_GIT_COMMIT
    [INFO] Building Docker Image zk_testbed
    ok



Figure:

    +-------------------------------------------------------------+
    |                                                             |
    |  +---------------+   +---------------+   +---------------+  |
    |  | Docker  (zk1) |   | Docker  (zk2) |   | Docker  (zk3) |  |
    |  +---------------+   +---------------+   +---------------+  |
    |          |                   |                   |          |
    |  +-------------------------------------------------------+  |
    |  |                Open vSwitch (and Ryu)                 |  |
    |  +-------------------------------------------------------+  |	
    |                              |                              |
    |  +-------------------------------------------------------+  |
    |  |                       Earthquake                      |  |
    |  +-------------------------------------------------------+  |	
    |                                                             |
    |                  Docker  (osrg/earthquake)                  |
    +-------------------------------------------------------------+


### Run Experiments

After you have set up the environment, you can run experiments as follows.
    
    docker$ ../../bin/earthquake run /tmp/zk-2212
    [INFO] Checking PYTHONPATH
    [INFO] Starting Earthquake Ethernet Switch
    [INFO] Switch PID: 28893
    [INFO] Starting Earthquake Ethernet Inspector
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

    docker$ ../../bin/earthquake tools summary /tmp/zk-2212
    Fri Jul 24 19:46:15 JST 2015 ...orage/naive/naive.go(142): a number of collected traces: 3
    00000002 caused failure

### [Experiment #0](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000): *not* reproduced the bug
zk2 was successfully promoted to an observer to a participant, because it received `UpToDate` before `Notification(config.version=100000000)`

* [32](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/32.event.json): zk1->zk2: `UpToDate`
* [36](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/36.event.json): zk1<-zk2: `FollowerInfo`
* [37](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/37.event.json): zk1<-zk2: `Notification(config.version=100000000)`
* [39](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/39.event.json): zk1->zk2: `Notification(config.version=100000000)`


zk3 was successfully promoted to an observer to a participant, because it received `UpToDate` before `Notification(config.version=100000000)`

* [20](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/20.event.json): zk1->zk3: `UpToDate`
* [21](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/21.event.json): zk1<-zk3: `Notification(config.version=100000000)`
* [23](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/23.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [25](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000000/actions/25.event.json): zk1<-zk3: `FollowerInfo`

### [Experiment #1](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001): *not* reproduced the bug
zk2 was already a participant when it received `UpToDate`

* [19](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001/actions/19.event.json): zk1<-zk2: `FollowerInfo`
* [31](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001/actions/31.event.json): zk1->zk2: `UpToDate`

zk3 was successfully promoted to an observer to a participant, because it received `UpToDate` before `Notification(config.version=100000000)`

* [26](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001/actions/26.event.json): zk1->zk3: `UpToDate`
* [29](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001/actions/29.event.json): zk1<-zk3: `FollowerInfo`
* [34](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001/actions/34.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [37](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000001/actions/37.event.json): zk1<-zk3: `Notification(config.version=100000000)`

### [Experiment #2](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000002): *reproduced* the bug (zk2, zk3)
zk2 was not able to be promoted (please see above for the reason)

* [12](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000002/actions/12.event.json): zk1->zk2: `Notification(config.version=100000000)`
* [19](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000002/actions/19.event.json): zk1->zk2: `UpToDate`
* No `FollowerInfo` (zk1<-zk2)

zk3 was not able to be promoted (please see above for the reason)

* [10](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000002/actions/10.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [24](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000002/actions/24.event.json): zk1->zk3: `UpToDate`
* No `FollowerInfo` (zk1<-zk3)

### [Experiment #3](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003): *reproduced* the bug (zk3)
zk2 was successfully promoted to an observer to a participant, because it received `UpToDate` before `Notification(config.version=100000000)`

* [24](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003/actions/24.event.json): zk1->zk2: `UpToDate`
* [25](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003/actions/25.event.json): zk1<-zk2: `Notification(config.version=100000000)`
* [27](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003/actions/27.event.json): zk1->zk2: `Notification(config.version=100000000)`
* [29](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003/actions/29.event.json): zk1<-zk2: `FollowerInfo`

zk3 was not able to be promoted (please see above for the reason)

* [7](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003/actions/7.event.json): zk1->zk3: `Notification(config.version=100000000)`
* [20](https://github.com/osrg/namazu/tree/v0.1.1/example/zk-found-2212.ryu/example-result.20150805/00000003/actions/20.event.json): zk1->zk3: `UpToDate`
* No `FollowerInfo` (zk1<-zk3)

## Conclusion
We found a distributed race condition bug of ZooKeeper, and identified its cause using Earthquake.

Through the experiments, we learned that the following points are important for implementing testing tool:

 * **Avoid false-positives**: i.e., the testing tool itself should not be bug-prone. False-positives complicates debugging. The authors of [MODIST](https://www.usenix.org/legacy/event/nsdi09/tech/full_papers/yang/yang_html/) \[nsdi09\] also alert this point.
 * **Don't modify the target software**: Modification complicates testing multiple versions of the target software. Hence it is hard to check whether the bug got fixed in new releases. Earthquake realizes non-invasive test by inspecting and reordering packets at Ethernet switch side.
 * **Support identifying the root cause of bugs**: Just finding bugs is not enough for improving quality of the target software. The quality gets improved only after identifying the root cause of bugs, and fixing them. Earthquake provides event history storage for estimating bug causes. We are also planning to add support for analyzing branch-coverage data (e.g. using JaCoCo) so as to pick up suspicious branch patterns.


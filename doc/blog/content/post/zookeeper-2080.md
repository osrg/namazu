+++
categories = ["blog"]
date = "2015-10-02"
tags = ["document","reproduced-bug","zookeeper"]
title = "ZOOKEEPER-2080: flaky JUnit test"

+++

## Introduction
[Apache Zookeeper](https://zookeeper.apache.org/) is a highly available coordination service that is used by many distributed systems.
(See also [our previous post about ZOOKEEPER-2212]({{< relref "post/zookeeper-2212.md" >}}))

ZooKeeper is comparatively matured, and well-tested with many JUnit scenarios.
However, due to its high non-determinism, some test scenarios are still flaky, plus it is very hard to identify the cause of such flaky bugs.

Using Earthquake, we successfully reproduced the bug [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080) and analyzed its cause.

## The Bug
[ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080) reports that the JUnit test `ReconfigRecoveryTest.testCurrentObserverIsParticipantInNewConfig` fails intermittently.

However, to the best of our knowledge, no one had successfully reproduced nor analyzed the bug for almost 2 years (since 2013), until we reproduced and analyzed with Earthquake.

We found that the bug has been caused by a race condition in the `QuorumCnxManager` class, which manages TCP sockets for leader election.

Unfortunately, to fix the bug gracefully, drastic changes to `QuorumCnxManager` (suggested since 2010: [ZOOKEEPER-901](https://issues.apache.org/jira/browse/ZOOKEEPER-901)) might be needed.
We continue to look on the progress of ZOOKEEPER-901 and several related tickets such as [ZOOKEEPER-900](https://issues.apache.org/jira/browse/ZOOKEEPER-900), [2164](https://issues.apache.org/jira/browse/ZOOKEEPER-2164) and [2246](https://issues.apache.org/jira/browse/ZOOKEEPER-2246).


## How to Reproduce the Bug with Earthquake

The bug can be easily reproduced by injecting several tens of millisecs sleeps to FLE (Fast Leader Election protocol) packets.
We did not have to permute packets explicitly.

NOTE: In this test case, Earthquake uses Linux Netfilter queue (NFQ) to hook loopback traffic of the pseudo-distributed cluster. Our NFQ library is available as [osrg/hookswitch](https://github.com/osrg/hookswitch/). HookSwitch can also hook OpenFlow packets, as [we used in ZOOKEEPER-2212]({{< relref "post/zookeeper-2212.md" >}}).

### Set up Earthquake (v0.1.2)
For more information about setting up Earthquake, please refer to [doc/how-to-setup-env.md](https://github.com/osrg.namazu/blob/v0.1.2/doc/how-to-setup-env.md).

    $ sudo sh -c 'echo 0 > /proc/sys/net/ipv4/tcp_autocorking' #recommended if Linux >= 3.14
    $ docker run --rm --tty --interactive --privileged -e EQ_DOCKER_PRIVILEGED=1 osrg.namazu:v0.1.2
    docker$ export PYTHONPATH=/earthquake #BUG(Dockerfile): should be predefined in > v0.1.2
    docker$ apt-get install ant-optional #BUG(Dockerfile): should be preinstalled in > v0.1.2
    docker$ cd /earthquake/example/zk-repro-2080.nfqhook
    docker$ ../../bin/earthquake init --force config.toml materials /tmp-zk-2080 #BUG(Dockerfile): in some envs, you may be unable to write in /tmp.

### Run Experiments
First, confirm that ZooKeeper works fine with*out* Earthquake (by setting `EQ_DISABLE`).

    docker$ EQ_DISABLE=1 ../../bin/earthquake run /tmp-zk-2080
    ..
    [junit] 2015-10-02 14:28:58,587 [myid:] - INFO  [main:ZKTestCase$1@65] - SUCCEEDED testCurrentObserverIsParticipantInNewConfig
    ..
    validation succeed


Then you can see that ZooKeeper almost always fails *with* Earthquake.

    docker$ ../../bin/earthquake run /tmp-zk-2080
    ..
    [junit] 2015-10-02 14:36:49,863 [myid:] - INFO  [main:ZKTestCase$1@70] - FAILED testCurrentObserverIsParticipantInNewConfig
    [junit] java.lang.AssertionError: waiting for server 2 being up
	..
    validation failed: exit status 1


If the bug could not be reproduced, you might have to modify the `sleep` parameter in `config.toml`. (about 30 msecs to 80 msecs)


### Analyze
Unlike we experienced in [ZOOKEEPER-2212]({{< relref "post/zookeeper-2212.md" >}}), neither Earthquake event history nor ZooKeeper logs were effective for analyzing cause of the bug.

Instead, we compared branch patterns of Java codes using our new Earthquake Analyzer.


    docker$ java -jar ../../bin/earthquake-analyzer.jar /tmp-zk-2080/ --classes-path /tmp-zk-2080/materials/zookeeper/build/classes
    [DEBUG] net.osrg.namazu.Analyzer - Scanning /tmp-zk-2080/00000000: experiment successful=true
    [DEBUG] net.osrg.namazu.Analyzer - Scanning /tmp-zk-2080/00000001: experiment successful=false
    ..
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection::getVote line 805-805
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection::lookForLeader line 919-919
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection::lookForLeader line 939-941
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection::lookForLeader line 945-945
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection::lookForLeader line 949-949
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection::lookForLeader line 951-952
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection$Messenger$WorkerReceiver::run line 393-395
    Suspicious: org.apache.zookeeper.server.quorum.FastLeaderElection$Messenger$WorkerReceiver::run line 403-404
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 808-809
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 812-814
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 816-816
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 818-818
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 823-825
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 830-830
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 833-834
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 837-839
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 870-870
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 878-880
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 882-882
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 884-884
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::queueCommittedProposals line 895-895
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::syncFollower line 744-746
    Suspicious: org.apache.zookeeper.server.quorum.LearnerHandler::syncFollower line 748-748
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::connectAll line 510-513
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::connectAll line 515-515
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::haveDelivered line 527-527
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::haveDelivered line 529-529
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager::receiveConnection line 382-382
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager$Listener::run line 657-657
    Suspicious: org.apache.zookeeper.server.quorum.QuorumCnxManager$SendWorker::run line 809-811
    

Earthquake Analyzer picks up some branch patterns peculiar to failed experiments, and marks them "suspicious".
Although the suspicious set can contain some misleading information, it is enough for approximation.

As mentioned above, the bug seems caused by a race condition between a TCP packet arrival and `SendWorker`/`RecvWorker` lifecycle.

Failed experiments tend to have a peculiar pattern of callling `SendWorker::finish()` (and also `RecvWorker::finish()`), e.g., calls from the branch [`QuorumCnxManager::receiveConnection line 382-382`](https://github.com/apache/zookeeper/blob/df7d56d25d38f872b5793af365ef732c4478eb1d/src/java/main/org/apache/zookeeper/server/quorum/QuorumCnxManager.java#L382).
	
When these peculiar calls to `SendWorker::finish()` are commented out, the bug gets hard to be reproduced. It suggests that `SendWorker`/`RecvWorker` lifecycles matters.

To fix the bug gracefully, using non-blocking TCP sockets (suggested in [ZOOKEEPER-901](https://issues.apache.org/jira/browse/ZOOKEEPER-901)) might be needed.

## Conclusion
We reproduced the flaky JUnit test bug [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080) and analyzed its cause using Earthquake.

Lessons we learned:

 * Just delaying packets is sometimes effective to reproduce flaky testcases.
 * Branch pattern analysis is an easy and general method for narrowing bug-cause candidates, although we still need much more improvements to this.

+++
date = "2015-08-20"
tags = ["document"]
title = "_index"

+++


# What is this?

In short, the goal of Earthquake project is providing a foundation of debugger for distributed systems.

Developing and maintaining distributed systems is difficult. 
The difficulty comes from many factors, 
but we believe that one of the most important reasons is lacking of a good debugger for distributed systems specific bugs.

[Read more..]({{< relref "about.md" >}})..

![Overview](/earthquake/images/overview.png)

# Found/Reproduced Bugs
* ZooKeeper:
 * Found [ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212) (race): [(blog article)]({{< relref "post/zookeeper-2212.md" >}})
 * Reproduced [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080) (race): [(blog article)]({{< relref "post/zookeeper-2080.md" >}})

* Etcd:
 * Found an etcd command line client (etcdctl) bug [#3517](https://github.com/coreos/etcd/issues/3517) (timing specification), fixed in [#3530](https://github.com/coreos/etcd/pull/3530). The fix also resulted a hint of [#3611](https://github.com/coreos/etcd/issues/3611): To Be Documented

* YARN:
 * Found [YARN-4301](https://issues.apache.org/jira/browse/YARN-4301) (fault tolerance): To Be Documented

The repro codes are located on [earthquake/example](https://github.com/osrg/earthquake/tree/master/example).

# How to use?
Please refer to [README file](https://github.com/osrg/earthquake/blob/master/README.md).

[This article]({{< relref "post/zookeeper-2212.md" >}}) is also a good start point.

# Contact
The project is managed on [github](https://github.com/osrg/earthquake).
[Pull requests](https://github.com/osrg/earthquake/pulls) and [issues](https://github.com/osrg/earthquake/issues) are welcome.
We are using [gitter](https://gitter.im/osrg/earthquake) for discussion.
Feel free to join.

[![Join the chat at https://gitter.im/osrg/earthquake](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/osrg/earthquake?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

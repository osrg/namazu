# Earthquake: Dynamic Model Checker for Distributed Systems

[![Join the chat at https://gitter.im/osrg/earthquake](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/osrg/earthquake?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Circle CI](https://circleci.com/gh/osrg/earthquake.svg?style=svg)](https://circleci.com/gh/osrg/earthquake)

Earthquake is a dynamic model checker (DMCK) for real implementations of distributed system (such as ZooKeeper).

[http://osrg.github.io/earthquake/](http://osrg.github.io/earthquake/)

Earthquakes permutes C/Java function calls, Ethernet packets, and injected fault events in various orders so as to find implementation-level bugs of the distributed system.
When Earthquake finds a bug, Earthquake automatically records [the event history](example/zk-found-2212.ryu) and helps you to analyze which permutation of events triggers the bug.

Basically, Earthquake permutes events in a random order, but you can write your [own state exploration policy](doc/arch.md) (in Go or Python) for finding deep bugs efficiently.

## News
 * Earthquake will be presented at the poster session of [ACM Symposium on Cloud Computing (SoCC)](http://acmsocc.github.io/2015/). (August 27-29, 2015, Hawaii)
 * We have successfully found a distributed race condition bug of ZooKeeper using Earthquake.
Please refer to [example/zk-found-2212.ryu](example/zk-found-2212.ryu) for further information.

## Quick Start
NOTE: [v0.1.1](https://github.com/osrg/earthquake/releases/tag/v0.1.1) might be stabler than the master branch.

 * How to set up the environment: [doc/how-to-setup-env.md](doc/how-to-setup-env.md)
 * Example: Finding a distributed race condition bug of ZooKeeper([ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212)): [example/zk-found-2212.ryu](example/zk-found-2212.ryu)

## Archtecture
Please see [doc/arch.md](doc/arch.md).

## How to Contribute
We welcome your contribution to Earthquake.
Please feel free to send your pull requests on github!

## Copyright
Copyright (C) 2015 [Nippon Telegraph and Telephone Corporation](http://www.ntt.co.jp/index_e.html).
Released under [Apache License 2.0](LICENSE).

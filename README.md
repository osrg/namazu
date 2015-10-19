# Earthquake: Dynamic Model Checker for Distributed Systems

[![Release](http://github-release-version.herokuapp.com/github/osrg/earthquake/release.svg?style=flat)](https://github.com/osrg/earthquake/releases/latest)
[![Join the chat at https://gitter.im/osrg/earthquake](https://img.shields.io/badge/GITTER-join%20chat-green.svg)](https://gitter.im/osrg/earthquake?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GoDoc](https://godoc.org/github.com/osrg/earthquake/earthquake?status.svg)](https://godoc.org/github.com/osrg/earthquake/earthquake)
[![Build Status](https://travis-ci.org/osrg/earthquake.svg?branch=master)](https://travis-ci.org/osrg/earthquake)

Earthquake is a dynamic model checker (DMCK) for real implementations of distributed system (such as ZooKeeper).

Blog: [http://osrg.github.io/earthquake/](http://osrg.github.io/earthquake/)

Earthquakes permutes C/Java function calls, Ethernet packets, Filesystem events, and injected faults in various orders so as to find implementation-level bugs of the distributed system.
When Earthquake finds a bug, Earthquake automatically records [the event history](example/zk-found-2212.ryu) and helps you to analyze which permutation of events triggers the bug.

Basically, Earthquake permutes events in a random order, but you can write your [own state exploration policy](doc/arch.md) (in Go or Python) for finding deep bugs efficiently.

## Found/Reproduced Bugs
 * ZooKeeper:
  * Found [ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212): [blog article](http://osrg.github.io/earthquake/post/zookeeper-2212/) ([repro code](example/zk-found-2212.ryu))
  * Reproduced [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080): [blog article](http://osrg.github.io/earthquake/post/zookeeper-2080/) ([repro code](example/zk-repro-2080.nfqhook))
 * etcd:
  * Found an etcd command line client (etcdctl) bug [#3517](https://github.com/coreos/etcd/issues/3517), fixed in [#3530](https://github.com/coreos/etcd/pull/3530): ([repro code](example/etcd/3517-reproduce)). The fix also resulted a hint of [#3611](https://github.com/coreos/etcd/pull/3611).

## Quick Start
 * How to set up the environment: [GettingStarted](http://osrg.github.io/earthquake/gettingStarted/) ([some extra info](doc/how-to-setup-env.md))
 * Example: Finding a distributed race condition bug of ZooKeeper([ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212)): [blog article](http://osrg.github.io/earthquake/post/zookeeper-2212/) ([repro code](example/zk-found-2212.ryu))

NOTE: [the latest release](https://github.com/osrg/earthquake/releases/latest) might be stabler than the master branch.

## Talks
 * Earthquake was presented at the poster session of [ACM Symposium on Cloud Computing (SoCC)](http://acmsocc.github.io/2015/). (August 27-29, 2015, Hawaii)

## How to Contribute
We welcome your contribution to Earthquake.
Please feel free to send your pull requests on github!

## Copyright
Copyright (C) 2015 [Nippon Telegraph and Telephone Corporation](http://www.ntt.co.jp/index_e.html).

Released under [Apache License 2.0](LICENSE).

---------------------------------------

## API Overview
```go
type MyPolicy struct {
	actionCh chan Action
}

func (p *MyPolicy) GetNextActionChan() chan Action {
	return p.actionCh
}

func (p *MyPolicy) QueueNextEvent(event Event) {
	// Possible events:
	//  - JavaFunctionEvent
	//  - PacketEvent
	//  - FilesystemEvent
	//  - LogEvent
	fmt.Printf("Event: %s\n", event)
	// You can also inject fault actions
	//  - PacketFaultAction
	//  - FilesystemFaultAction
	//  - ShellAction
	action, err := event.DefaultAction()
	if err != nil {
		panic(err)
	}
	// send in a goroutine so as to make the function non-blocking.
	go func() {
		fmt.Printf("Action ready: %s\n", action)
		p.actionCh <- action
		fmt.Printf("Action passed: %s\n", action)
	}()
}

func NewMyPolicy() ExplorePolicy {
	return &MyPolicy{actionCh: make(chan Action)}
}

func main(){
	RegisterPolicy("mypolicy", NewMyPolicy)
	os.Exit(CLIMain(os.Args))
}
```
Please refer to [example/template](example/template) for further information.

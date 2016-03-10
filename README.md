# Earthquake: Programmable Fuzzy Scheduler for Testing Distributed Systems

[![Release](http://github-release-version.herokuapp.com/github/osrg/earthquake/release.svg?style=flat)](https://github.com/osrg/earthquake/releases/latest)
[![Join the chat at https://gitter.im/osrg/earthquake](https://img.shields.io/badge/GITTER-join%20chat-green.svg)](https://gitter.im/osrg/earthquake?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GoDoc](https://godoc.org/github.com/osrg/earthquake/earthquake?status.svg)](https://godoc.org/github.com/osrg/earthquake/earthquake)
[![Build Status](https://travis-ci.org/osrg/earthquake.svg?branch=master)](https://travis-ci.org/osrg/earthquake)
[![Coverage Status](https://coveralls.io/repos/github/osrg/earthquake/badge.svg?branch=master)](https://coveralls.io/github/osrg/earthquake?branch=master)

Earthquake is a programmable fuzzy scheduler for testing real implementations of distributed system (such as ZooKeeper).

Blog: [http://osrg.github.io/earthquake/](http://osrg.github.io/earthquake/)

Earthquakes permutes C/Java function calls, Ethernet packets, Filesystem events, and injected faults in various orders so as to find implementation-level bugs of the distributed system.
Earthquake can also control non-determinism of the thread interleaving (by calling `sched_setattr(2)` with randomized parameters).
So Earthquake can be also used for testing standalone multi-threaded software.

Basically, Earthquake permutes events in a random order, but you can write your [own state exploration policy](doc/arch.md) (in Golang) for finding deep bugs efficiently.

## Found/Reproduced Bugs
 * ZooKeeper:
  * Found [ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212) (race): [blog article](http://osrg.github.io/earthquake/post/zookeeper-2212/) ([repro code](example/zk-found-2212.ryu))
  * Reproduced [ZOOKEEPER-2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080) (race): [blog article](http://osrg.github.io/earthquake/post/zookeeper-2080/) ([repro code](example/zk-repro-2080.nfqhook))
 * etcd:
  * Found an etcd command line client (etcdctl) bug [#3517](https://github.com/coreos/etcd/issues/3517) (timing specification), fixed in [#3530](https://github.com/coreos/etcd/pull/3530): ([repro code](example/etcd/3517-reproduce)). The fix also resulted a hint of [#3611](https://github.com/coreos/etcd/pull/3611).
  * Reproduced flaky tests {[#4006](https://github.com/coreos/etcd/pull/4006), [#4039](https://github.com/coreos/etcd/issues/4039)} ([repro instruction](http://www.slideshare.net/AkihiroSuda/tackling-nondeterminism-in-hadoop-testing-and-debugging-distributed-systems-with-earthquake-57866497/42))
 * YARN:
  * Found [YARN-4301](https://issues.apache.org/jira/browse/YARN-4301) (fault tolerance): ([repro code](example/yarn/4301-reproduce))
  * Reproduced flaky tests YARN-{[1978](https://issues.apache.org/jira/browse/YARN-1978), [4168](https://issues.apache.org/jira/browse/YARN-4168), [4543](https://issues.apache.org/jira/browse/YARN-4543), [4548](https://issues.apache.org/jira/browse/YARN-4548), [4556](https://issues.apache.org/jira/browse/YARN-4556)} ([repro instruction](http://www.slideshare.net/AkihiroSuda/tackling-nondeterminism-in-hadoop-testing-and-debugging-distributed-systems-with-earthquake-57866497/42))

## Quick Start
The following instruction shows how you can start *Earthquake Container*, the simplified CLI for Earthquake.


    $ sudo apt-get install libzmq3-dev libnetfilter-queue-dev
    $ go get github.com/osrg/earthquake/earthquake-container
    $ sudo earthquake-container run -it --rm ubuntu bash


In *Earthquake Container*, you can run arbitrary command that might be *flaky*.
JUnit tests are interesting to try.

    earthquake-container$ git clone something
    earthquake-container$ cd something
    earthquake-container$ for f in $(seq 1 1000);do mvn test; done


You can also specify a config file (`--eq-config` option for `earthquake-container`.)
A typical configuration file (`config.toml`) is as follows:

```toml
# Policy for observing events and yielding actions
# You can also implement your own policy.
# Default: "random"
explorePolicy = "random"

[explorePolicyParam]
  # for Ethernet/Filesystem/Java inspectors, event are non-deterministically delayed.
  # minInterval and maxInterval are bounds for the non-deterministic delays
  # Default: 0 and 0
  minInterval = "80ms"
  maxInterval = "3000ms"
  procPolicy = "dirichlet"

[explorePolicyParam.procParam]
  # resetProbability is a probability for resetting process scheduling attributes (for Process inspector. procPolicy="dirichlet")
  # Default: 0.1 (10%)
  procResetSchedProbability = 0.1

[container]
  # Default: false
  enableEthernetInspector = true
  # Default: true
  enableProcInspector = true
  # Default: "1s"
  procWatchInterval = "1s"
```
For other parameters, please refer to [`config.go`](earthquake/util/config/config.go) and [`randompolicy.go`](earthquake/explorepolicy/random/randompolicy.go).

If you don't want to use containers, you can also use Earthquake with an arbitrary process tree.

    $ go get github.com/osrg/earthquake/earthquake
    $ sudo earthquake inspectors proc -root-pid $TARGET_PID -watch-interval 1s -autopilot config.toml

For full-stack (fully-distributed) Earthquake environment, please refer to [doc/how-to-setup-env-full.md](doc/how-to-setup-env-full.md).

[The slides for the presentation at FOSDEM](http://www.slideshare.net/AkihiroSuda/tackling-nondeterminism-in-hadoop-testing-and-debugging-distributed-systems-with-earthquake-57866497/42) might be also helpful.

## Talks

 * [FOSDEM](https://fosdem.org/2016/schedule/event/nondeterminism_in_hadoop/) (January 30-31, 2016, Brussels)
 * The poster session of [ACM Symposium on Cloud Computing (SoCC)](http://acmsocc.github.io/2015/) (August 27-29, 2015, Hawaii)

## How to Contribute
We welcome your contribution to Earthquake.
Please feel free to send your pull requests on github!

## Copyright
Copyright (C) 2015 [Nippon Telegraph and Telephone Corporation](http://www.ntt.co.jp/index_e.html).

Released under [Apache License 2.0](LICENSE).

---------------------------------------

## API Overview
```go
// implements earthquake/explorepolicy/ExplorePolicy interface
type MyPolicy struct {
	actionCh chan Action
}

func (p *MyPolicy) ActionChan() chan Action {
	return p.actionCh
}

func (p *MyPolicy) QueueEvent(event Event) {
	// Possible events:
	//  - JavaFunctionEvent (byteman)
	//  - PacketEvent (Netfilter, Openflow)
	//  - FilesystemEvent (FUSE)
	//  - ProcSetEvent (Linux procfs)
	//  - LogEvent (syslog)
	fmt.Printf("Event: %s\n", event)
	// You can also inject fault actions
	//  - PacketFaultAction
	//  - FilesystemFaultAction
	//  - ProcSetSchedAction
	//  - ShellAction
	action, err := event.DefaultAction()
	if err != nil {
		panic(err)
	}
	// send in a goroutine so as to make the function non-blocking.
	// (Note that earthquake/util/queue/TimeBoundedQueue provides
	// better semantics and determinism, this is just an example.)
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

## Known Limitation
After running Earthquake (process inspector) many times, `sched_setattr(2)` can fail with `EBUSY`.
This seems to be a bug of kernel; We're looking into this.

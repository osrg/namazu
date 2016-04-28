+++
categories = ["blog"]
date = "2016-02-05"
tags = ["document"]
title = "FOSDEM: Thank you for attending our presentation!"

+++

On January 30, Brussels, we had a very good time at [FOSDEM 2016](https://fosdem.org/2016/schedule/event/nondeterminism_in_hadoop/)!

Thank you all for attending our presentation and discussing with us!

![fosdem2016_01.jpg](/namazu/images/fosdem2016_01.jpg)

{{< slideshare key="MzuiqJw0xFIpo8" slide="1" >}}


At the presentation, we showed some issues we found or reproduced: ZOOKEEPER-{[2080](https://issues.apache.org/jira/browse/ZOOKEEPER-2080),[2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212)}, YARN-{[1978](https://issues.apache.org/jira/browse/YARN-1978),[4168](https://issues.apache.org/jira/browse/YARN-4168),[4543](https://issues.apache.org/jira/browse/YARN-4543),[4548](https://issues.apache.org/jira/browse/YARN-4548),[4556](https://issues.apache.org/jira/browse/YARN-4556)}, and etcd {[#4006](https://github.com/coreos/etcd/pull/4006), [#4039](https://github.com/coreos/etcd/issues/4039)}.

Recently, the number of these reproducible issues increased significantly due to the new feature we started to work on: [**the thread scheduling fuzzer**](http://www.slideshare.net/AkihiroSuda/tackling-nondeterminism-in-hadoop-testing-and-debugging-distributed-systems-with-earthquake-57866497/32).

This feature fuzzes the thread scheduling for a specific process tree by using Linux (>= 3.14) `SCHED_DEADLINE` scheduler and calling [`sched_setattr(2)`](http://man7.org/linux/man-pages/man2/sched_getattr.2.html) with randomized parameters.

The feature is planned to be included in the next release of Earthquake (v0.2.0). However, [the early preview release is already available here](https://github.com/AkihiroSuda/MicroEarthquake/tree/v20160203).

Quick start slides are also available:
{{< slideshare key="MzuiqJw0xFIpo8" slide="42" >}}

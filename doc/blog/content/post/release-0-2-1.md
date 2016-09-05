+++
categories = ["blog"]
date = "2016-09-05"
tags = ["document"]
title = "Release Namazu v0.2.1, with Namazu Swarm v0.0.1!"

+++

We are glad to annouce the release of [Namazu](https://github.com/osrg/namazu) v0.2.1.

![Overview](/namazu/images/namazu-v0.2.png)

Namazu v0.2.1 is a maintanance release of [Namazu v0.2.0]({{< relref "post/release-0-2-0.md" >}})

You can download the Namazu v0.2.1 binary release from [github](https://github.com/osrg/namazu/releases/tag/v0.2.1).

Or you can also build Namazu manually:

    $ sudo apt-get install libzmq3-dev libnetfilter-queue-dev
    $ go get github.com/osrg/namazu/nmz

## Changes from v0.2.0

 * #167, #168, #169, #170: doc: miscellaneous improvements
 * #166: vendor go packages
 * #163: *: bump up Go to 1.7
 * #162: container: add support for inspectors/fs
 * #160: inspectors/fs: implement Fsync
 * #158, #159: inspectors/fs: improved CLI (thank you @v01dstar !)
 * #156: inspectors/proc: improved error handling
 * #154, #155: inspectors/proc: improved CLI

## Namazu Swarm v0.0.1

We also released the first version of [Namazu Swarm](https://github.com/osrg/namazu-swarm), CI Job Parallelizer built on Docker and Kubernetes
[Namazu Swarm](https://github.com/osrg/namazu-swarm) is developed as a part of Namazu, but it does not depends on Namazu (although you can combine them together).

![Namazu Swarm](https://raw.githubusercontent.com/osrg/namazu-swarm/507f1ea51790ebc6d64740e8eb14e009d0353970/docs/img/nmzswarm.png)

Namazu Swarm is hosted at [osrg/namazu-swarm](https://github.com/osrg/namazu-swarm).

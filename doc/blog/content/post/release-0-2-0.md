+++
categories = ["blog"]
date = "2016-05-20"
tags = ["document"]
title = "Release Namazu v0.2.0"

+++
Note: we recently renamed _Earthquake_ to _Namazu_.

We are glad to annouce the release of [Namazu](https://github.com/osrg/namazu) v0.2.0.

![Overview](/namazu/images/namazu-v0.2.png)

Namazu v0.2.0 includes many new features: Process inspector, Filesystem inspector, Container CLI, Semi-deterministic replaying API...

These new features made Namazu much more powerful and simple.
Now, no configuration is needed to get started with Namazu.

You can download the Namazu v0.2.0 binary release from [github](https://github.com/osrg/namazu/releases/tag/v0.2.0).

Or you can also build Namazu manually:

    $ sudo apt-get install libzmq3-dev libnetfilter-queue-dev
    $ go get github.com/osrg/namazu/nmz

## New features
### Process inspector
The process inspector sets random scheduling priority to threads under a specific Linux process:

    $ sudo nmz inspectors proc -pid $TARGET_PID -watch-interval 1s

The process inspector is sometimes useful when you want to reproduce flaky xUnit tests.

The experimental result for Hadoop tests is available in the slide we presented at ApacheCon:
{{< slideshare key="esat324HuI0vud" slide="41" >}}

### Filesystem inspector

The filesystem inspector provides randomized scheduling and fault injection for filesystem using FUSE:

    $ mkdir /tmp/{nmzfs-orig,nmzfs}
    $ sudo nmz inspectors fs -original-dir /tmp/nmzfs-orig -mount-point /tmp/nmzfs -autopilot config.toml
	$ $TARGET_PROGRAM_WHICH_ACCESSES_TMP_NMZFS
	$ sudo fusermount -u /tmp/nmzfs

Using filesystem inspector, we successfully found [YARN-4301](https://issues.apache.org/jira/browse/YARN-4301).

{{< slideshare key="esat324HuI0vud" slide="55" >}}

### Container CLI

We introduced *Namazu Container*, a new human-friendly, Docker-like CLI:

    $ sudo nmz container run -it --rm -v /foo:/foo ubuntu bash

In *Namazu Container*, you can run arbitrary command that might be *flaky*.
JUnit tests are interesting to try.

    nmzc$ git clone something
    nmzc$ cd something
    nmzc$ for f in $(seq 1 1000);do mvn test; done


By default, only process inspector is enabled in *Namazu Container*.
Please refer to [README file](https://github.com/osrg/namazu/blob/master/README.md) for configuration.

### Semi-deterministic replaying API

Semi-deterministic replayer is an experimental feature:
it determines a delay for an event using a seed value and the hash of the event, rather than just using a random value.

It does not guarantee full determinism, but we believe it is sometimes enough for debugging.

{{< slideshare key="esat324HuI0vud" slide="58" >}}

### Miscellaneous improvements

 * Renamed Earthquake to Namazu, and introduced a logo image
 * Support static build
 * Unit tests for Namazu itself

## Talks
Recently, we made presentations at two events: [CoreOS Fest](http://sched.co/6Szb) (May 10) and [ApacheCon Core NA](http://sched.co/6OJU) (May 12).

Please refer to the [article]({{< relref "post/coreosfest2016-and-apachecon2016.md" >}}) in our blog.

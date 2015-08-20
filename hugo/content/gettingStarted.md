+++
categories = ["general"]
date = "2015-08-20"
tags = ["document"]
title = "Getting Started"

+++
All you have to do is make Docker installed on your host and run the pre-built Docker image [osrg/earthquake](https://registry.hub.docker.com/u/osrg/earthquake/).

[![Docker Hub](http://dockeri.co/image/osrg/earthquake)](https://registry.hub.docker.com/u/osrg/earthquake/)

    
    $ docker run --rm --tty --interactive osrg/earthquake
	INIT: Running without privileged mode. Please set EQ_DOCKER_PRIVILEGED if you want to use Ethernet Inspector
	INIT: Earthquake is installed on /earthquake. Please refer to /earthquake/README.md
	INIT: Starting command: ['/bin/bash', '--login', '-i']
	root@a0c2e4413483:/earthquake# ^D
	INIT: Exiting with status 0..(['/bin/bash', '--login', '-i'])
	
Then, you can do the things what you want in `/earthquake` directory.
You might want to try several [examples](https://github.com/osrg/earthquake/blob/master/example).


## Privileged Mode (provides Docker-in-Docker, Open vSwitch, and Ryu)
This mode might be useful for Ethernet Inspector.
    
    $ sudo modprobe openvswitch
    $ docker run --rm --tty --interactive --privileged -e EQ_DOCKER_PRIVILEGED=1 osrg/earthquake 
    INIT: Running with privileged mode. Enabling DinD, OVS, and Ryu
    INIT: Earthquake is installed on /earthquake. Please refer to /earthquake/README.md
    INIT: Starting command: ['wrapdocker', '/init.dind-ovs-ryu.sh']
    * /etc/openvswitch/conf.db does not exist
    * Creating empty database /etc/openvswitch/conf.db
    * Starting ovsdb-server
    * Configuring Open vSwitch system IDs
    * Starting ovs-vswitchd
    * Enabling remote OVSDB managers
    Assigned 192.168.42.254 to ovsbr0
    root@907529be8b21:/earthquake# ^D
	INIT: Exiting with status 0..(['wrapdocker', '/init.dind-ovs-ryu.sh'])


[README file](https://github.com/osrg/earthquake/blob/master/README.md) and [this article]({{< relref "post/zookeeper-2212.md" >}}) are also good start points.


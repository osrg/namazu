# How to setup the environment for Earthquake

All you have to do is make Docker installed on your host and run the pre-built Docker image [osrg/earthquake](https://registry.hub.docker.com/u/osrg/earthquake/).

[![Docker Hub](http://dockeri.co/image/osrg/earthquake)](https://registry.hub.docker.com/u/osrg/earthquake/)

    
    $ docker run --rm --tty --interactive osrg/earthquake
	INIT: Running without privileged mode. Please set EQ_DOCKER_PRIVILEGED if you want to use Ethernet Inspector
	INIT: Earthquake is installed on /earthquake. Please refer to /earthquake/README.md
	INIT: Starting command: ['/bin/bash', '--login', '-i']
	root@a0c2e4413483:/earthquake# ^D
	INIT: Exiting with status 0..(['/bin/bash', '--login', '-i'])
	
Then, you can do the things what you want in `/earthquake` directory.
You might want to try several [examples](../example).


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
    

## Set up the environment manually (You might NOT need to read this section)
If you want to set up your own environment manually, please follow this instruction.

	$ sudo apt-get install -y --no-install-recommends protobuf-compiler default-jdk maven
    $ curl https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz | sudo tar Cxz /usr/local
    $ export PATH=/usr/local/go/bin:$PATH 

NOTE: Go 1.5 or later is required to build libearthquake.so.

### (Optional) Install Dependencies for MongoDB history storage
    
    $ sudo apt-get install -y --no-install-recommends mongodb
    
### (Optional) Install Dependencies for pyearthquake
    
	$ sudo apt-get install -y --no-install-recommends python-{flask,scapy,zmq}
    $ sudo pip install hexdump
    
### (Optional) Install Dependencies for Ethernet inspector (ryu)
#### Install Open vSwitch
    
    $ sudo apt-get install -y --no-install-recommends openvswitch-switch
    
Earthquake is tested with Open vSwitch 2.3.1 (Ubuntu 15.04).

#### Install ryu
    
    $ sudo pip install ryu
    
Earthquake is tested with ryu 3.20.2

#### Install pipework
    
    $ sudo apt-get install -y --no-install-recommends arping
    $ sudo curl https://raw.githubusercontent.com/jpetazzo/pipework/master/pipework -o /usr/local/bin/pipework
    $ sudo chmod +x /usr/local/bin/pipework
    

#### Set up Open vSwitch
	
    $ sudo ovs-vsctl add-br ovsbr0
    $ sudo ovs-vsctl set bridge ovsbr0 protocols=OpenFlow13
    $ sudo ovs-vsctl set-controller ovsbr0 tcp:127.0.0.1
    $ sudo echo 'ip addr add 192.168.42.254/24 dev ovsbr0' > /etc/rc.local
    $ sudo sh /etc/rc.local

### (Optional) Install Dependencies for Ethernet inspector (nfqhook)
    
    $ sudo apt-get install -y --no-install-recommends libnetfilter-queue1 python-prctl
    
### Build Earthquake
    
    $ cd /path/to/earthquake
    $ ./build
    

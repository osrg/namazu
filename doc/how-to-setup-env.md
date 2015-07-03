# How to setup the environment for Earthquake

All you have to do is make Docker installed on your host and run the pre-built Docker image [osrg/earthquake](https://registry.hub.docker.com/u/osrg/earthquake/).

[![Docker Hub](http://dockeri.co/image/osrg/earthquake)](https://registry.hub.docker.com/u/osrg/earthquake/)

    
    $ docker run --privileged -t -i osrg/earthquake
	* /etc/openvswitch/conf.db does not exist
	* Creating empty database /etc/openvswitch/conf.db
	* Starting ovsdb-server
	* Configuring Open vSwitch system IDs
	* Starting ovs-vswitchd
	* Enabling remote OVSDB managers
	Assigned 192.168.42.254 to ovsbr0
	root@48ce4641421a:/earthquake#
    
Then, you can do the things what you want in `/earthquake` directory.
You might want to try several [examples](../example).

Notes:

 * If you want to use Ethernet inspector, you should load  Open vSwitch kernel module (`sudo modprobe openvswitch`) before running the Docker image.
 * `--privileged` is required if you want to use Ethernet inspector or Docker-in-Docker.
    

## Set up the environment manually (You might NOT need to read this section)
If you want to set up your own environment manually, please follow this instruction.

### Install Go
You need to install Go 1.5 or later to build libearthquake.so.
    
    $ git clone -b release-branch.go1.4 https://go.googlesource.com/go $HOME/go1.4
    $ (cd $HOME/go1.4/src; ./make.bash) # golang 1.4 (in $HOME/go1.4/bin) is required to build golang 1.5
    $ git clone https://go.googlesource.com/go $HOME/go
    $ (cd $HOME/go/src; ./make.bash)
    $ export PATH=$HOME/go/bin:$PATH 


### Install JDK

    $ sudo apt-get install -y default-jdk maven

### Install Dependencies for pyearthquake
    
    $ sudo pip install flask scapy zmq hexdump
    
### Install Dependencies for Ethernet inspector (ryu)
#### Install Open vSwitch
    
    $ sudo apt-get install -y openvswitch-switch
    
Earthquake is tested with Open vSwitch 2.3.1 (Ubuntu 15.04).

#### Install ryu
    
    $ sudo pip install ryu
    
Earthquake is tested with ryu 3.20.2

#### Install pipework
    
    $ sudo apt-get install -y arping
    $ sudo wget --no-check-certificate --quiet https://raw.githubusercontent.com/jpetazzo/pipework/master/pipework -O /usr/local/bin/pipework
    $ sudo chmod +x /usr/local/bin/pipework
    

#### Set up Open vSwitch
	
    $ sudo ovs-vsctl add-br ovsbr0
    $ sudo ovs-vsctl set bridge ovsbr0 protocols=OpenFlow13
    $ sudo ovs-vsctl set-controller ovsbr0 tcp:127.0.0.1
    $ sudo echo 'ip addr add 192.168.42.254/24 dev ovsbr0' > /etc/rc.local
    $ sudo sh /etc/rc.local

### Install Dependencies for Ethernet inspector (nfqhook)
    
    $ sudo apt-get install -y libnetfilter-queue1 python-prctl
    

### Build Earthquake
    
    $ cd /path/to/earthquake
    $ ./build
    

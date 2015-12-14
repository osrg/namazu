# How to setup the environment for Earthquake (Full stack)

All you have to do is make Docker installed on your host and run the pre-built Docker image [osrg/earthquake](https://registry.hub.docker.com/u/osrg/earthquake/).

[![Docker Hub](http://dockeri.co/image/osrg/earthquake)](https://registry.hub.docker.com/u/osrg/earthquake/)

    
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
    root@907529be8b21:/earthquake# 

Then, you can do the things what you want in `/earthquake` directory.
You might want to try several [examples](../example).

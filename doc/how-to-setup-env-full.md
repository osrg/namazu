# How to setup the environment for Namazu (Full stack)

All you have to do is make Docker installed on your host and run the pre-built Docker image [osrg/namazu](https://registry.hub.docker.com/u/osrg/namazu/).

[![Docker Hub](http://dockeri.co/image/osrg/namazu)](https://registry.hub.docker.com/u/osrg/namazu/)

    
    $ sudo modprobe openvswitch
    $ docker run --rm --tty --interactive --privileged -e NMZ_DOCKER_PRIVILEGED=1 osrg/namazu 
    INIT: Running with privileged mode. Enabling DinD, OVS, and Ryu
    INIT: Namazu is installed on /namazu. Please refer to /namazu/README.md
    INIT: Starting command: ['wrapdocker', '/init.dind-ovs-ryu.sh']
    * /etc/openvswitch/conf.db does not exist
    * Creating empty database /etc/openvswitch/conf.db
    * Starting ovsdb-server
    * Configuring Open vSwitch system IDs
    * Starting ovs-vswitchd
    * Enabling remote OVSDB managers
    Assigned 192.168.42.254 to ovsbr0
    root@907529be8b21:/namazu# 

Then, you can do the things what you want in `/namazu` directory.
You might want to try several [examples](../example).

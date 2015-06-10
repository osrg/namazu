# How to Install Dependencies
## Install OVS
    
    # apt-get install -y openvswitch-switch # tested with OVS 2.3.1 (Ubuntu 15.04)

## Install Python packages
Some of them are needed by ryu, not by earthquake itself.
    
    # apt-get install -y python-{colorama,dev,eventlet,flask,lxml,msgpack,netaddr,networkx,oslo.config,paramiko,pip,routes,scapy,six,webob,zmq}
    
## Install Ryu
    
    # pip install ryu==3.20.2
    
## Install pipework
    
    # apt-get install -y arping
    # wget --no-check-certificate --quiet https://raw.githubusercontent.com/AkihiroSuda/pipework/fix-pid-conflict/pipework -O /usr/local/bin/pipework
    # chmod +x /usr/local/bin/pipework
    
## Setup OVS
    
    # ovs-vsctl add-br ovsbr0
    # ovs-vsctl set bridge ovsbr0 protocols=OpenFlow13
    # ovs-vsctl set-controller ovsbr0 tcp:127.0.0.1
    # echo 'ip addr add 192.168.42.254/24 dev ovsbr0' > /etc/rc.local
    # sh /etc/rc.local
    

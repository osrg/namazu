#!/usr/bin/env python
# NOTE: run with ryu-manager

## FIXME: move these ones to the config file
TCP_PORTS=[2181, 2888, 3888]
ZMQ_ADDR='ipc:///tmp/eq/ether_inspector'

from pyearthquake.middlebox.switch import RyuOF13Switch

class ZkSwitch(RyuOF13Switch):
    def __init__(self, *args, **kwargs):
        super(ZkSwitch, self).__init__( tcp_ports=TCP_PORTS, 
                                                udp_ports=[], 
                                                zmq_addr=ZMQ_ADDR, 
                                                *args, **kwargs )

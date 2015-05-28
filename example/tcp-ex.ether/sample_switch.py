#!/usr/bin/env python
# NOTE: run with ryu-manager

## FIXME: move these ones to the config file
TCP_PORTS=[9999]
ZMQ_ADDR='ipc:///tmp/eq/ether_inspector'

from pyearthquake.middlebox.switch import RyuOF13Switch

class SampleOF13Switch(RyuOF13Switch):
    def __init__(self, *args, **kwargs):
        super(SampleOF13Switch, self).__init__( tcp_ports=TCP_PORTS, 
                                                udp_ports=[], 
                                                zmq_addr=ZMQ_ADDR, 
                                                *args, **kwargs )

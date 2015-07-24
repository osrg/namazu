#!/usr/bin/env python
# NOTE: run with ryu-manager
import os

ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')

from pyearthquake.middlebox.switch import RyuOF13Switch


class Zk2212Switch(RyuOF13Switch):
    def __init__(self, *args, **kwargs):
        super(Zk2212Switch, self).__init__(tcp_ports=[2181, 2888, 3888],
                                           udp_ports=[],
                                           zmq_addr=ZMQ_ADDR,
                                           *args, **kwargs)

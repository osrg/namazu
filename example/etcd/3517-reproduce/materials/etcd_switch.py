#!/usr/bin/env python
# NOTE: run with ryu-manager
import os

ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')

from pyearthquake.middlebox.switch import RyuOF13Switch

class EtcdSwitch(RyuOF13Switch):
    def __init__(self, *args, **kwargs):
        super(EtcdSwitch, self).__init__(tcp_ports=[4001, 7001],
                                           udp_ports=[],
                                           zmq_addr=ZMQ_ADDR,
                                           *args, **kwargs)


#!/usr/bin/env python
import os

ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')
import pyearthquake
from pyearthquake.inspector.zookeeper import ZkEtherInspector

LOG = pyearthquake.LOG.getChild(__name__)

class Zk2080Inspector(ZkEtherInspector):

    def __init__(self, zmq_addr):
        super(Zk2080Inspector, self).__init__(zmq_addr,
                                              ignore_pings=True,
                                              dump_bad_packet=False)

if __name__ == '__main__':
    d = Zk2080Inspector(zmq_addr=ZMQ_ADDR)
    d.start()

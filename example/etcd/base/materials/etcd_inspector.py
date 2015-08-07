#!/usr/bin/env python
import os
from pyearthquake.signal.event import PacketEvent
import base64

ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')

import pyearthquake
from pyearthquake.inspector.ether import EtherInspectorBase

LOG = pyearthquake.LOG.getChild(__name__)


class EtcdInspector(EtherInspectorBase):

    def __init__(self, zmq_addr):
        super(EtcdInspector, self).__init__(zmq_addr)

    def map_packet_to_event(self, pkt):
        return PacketEvent.from_message(src_entity="dummy", dst_entity="dummy", message=base64.b64encode(str((pkt))))

if __name__ == '__main__':
    print ''
    d = EtcdInspector(zmq_addr=ZMQ_ADDR)
    d.start()

#!/usr/bin/env python
import os
from pyearthquake.signal.event import PacketEvent
import base64

from scapy.all import *
from scapy.layers import http

ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')

import pyearthquake
from pyearthquake.inspector.ether import EtherInspectorBase

LOG = pyearthquake.LOG.getChild(__name__)

class EtcdInspector(EtherInspectorBase):

    def __init__(self, zmq_addr):
        super(EtcdInspector, self).__init__(zmq_addr)
        self.regist_layer_on_tcp(scapy.layers.http.HTTP, 7001)

    def map_packet_to_event(self, pkt):
        if not "IP" in pkt:
            return None
        ipPayload = pkt['IP']

        if not "HTTP" in pkt:
            return None

        httpPaylod = str(pkt['HTTP'])
        return PacketEvent.from_message(src_entity=ipPayload.src, dst_entity=ipPayload.dst, message=base64.b64encode(str((httpPaylod))))

if __name__ == '__main__':
    print ''
    d = EtcdInspector(zmq_addr=ZMQ_ADDR)
    d.start()

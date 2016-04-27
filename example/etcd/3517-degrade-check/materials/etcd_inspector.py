#!/usr/bin/env python
import os
from pynmz.signal.event import PacketEvent
import base64

from scapy.all import *

ZMQ_ADDR = os.getenv('NMZ_ETHER_ZMQ_ADDR')

import pynmz
from pynmz.inspector.ether import EtherInspectorBase

LOG = pynmz.LOG.getChild(__name__)

class EtcdPacket(Packet):
    name = 'EtcdPacket'
    longname = 'EtcdPacket'
    # fields_desc=[ StrFixedLenField('type', '[NUL]', 5),
    #               StrStopField('msg', '(NULL)', '\r\n') ]
    
    def post_dissect(self, s):
        try:
            msg = {'asdf': 'hjkl'}
            src_entity = 'server'
            dst_entity =  'client'
            self.event = PacketEvent.from_message(src_entity, dst_entity, msg)
        except Exception as e:
            LOG.exception(e)
            
    def mysummary(self):
        """
        human-readable summary
        """
        try:
            msg = self.event.option['message']
            src_entity = self.event.option['src_entity']
            dst_entity = self.event.option['dst_entity']
            return self.sprintf('%s ==> %s EtcdPacket msg: %s' % \
                                (src_entity, dst_entity, msg['asdf']))
        except Exception as e:
            LOG.exception(e)
            return self.sprintf('ERROR')

class EtcdInspector(EtherInspectorBase):

    def __init__(self, zmq_addr):
        super(EtcdInspector, self).__init__(zmq_addr)
        self.regist_layer_on_tcp(EtcdPacket, 7001)

    def map_packet_to_event(self, pkt):
        return PacketEvent.from_message(src_entity="dummy", dst_entity="dummy", message=base64.b64encode(str((pkt))))
        # if pkt.haslayer(EtcdPacket):
        #     LOG.info('%s packet: %s', self.__class__.__name__, pkt[EtcdPacket].mysummary())            
        #     event = pkt[EtcdPacket].event
        #     LOG.info('mapped event=%s', event)
        #     return event
        # else:
        #     LOG.info('%s unknown packet: %s', self.__class__.__name__, pkt.mysummary())
        #     # hexdump.hexdump(str(pkt))
        #     return None
        # return PacketEvent.from_message(src_entity="dummy", dst_entity="dummy", message=base64.b64encode(str(pkt)))

if __name__ == '__main__':
    print ''
    d = EtcdInspector(zmq_addr=ZMQ_ADDR)
    d.start()

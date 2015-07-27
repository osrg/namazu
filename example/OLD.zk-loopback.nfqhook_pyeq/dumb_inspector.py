#!/usr/bin/env python

## FIXME: move these ones to the config file
ZMQ_ADDR='ipc:///tmp/eq/ether_inspector'

import colorama
from scapy.all import *
import pyearthquake
from pyearthquake.inspector.ether import *
from pyearthquake.signal.signal import *
from pyearthquake.signal.event import *
from pyearthquake.signal.action import *
import hexdump as hd # hexdump conflict with scapy.all.hexdump


LOG = pyearthquake.LOG.getChild(__name__)

# terrible table (TODO: move to config.json)
pid_table = {}

class Util(object):
    ## scapy hates inheritance from scapy.all.Packet class

    @classmethod
    def ip2pid(cls, ip):
        # TODO: move to pyearthquake.util
        try:
            return pid_table[ip]
        except KeyError:
            return '_unknown_process_%s' % (ip)


    @classmethod
    def make_event(cls, klazz):
        tcp, ip = klazz.underlayer, klazz.underlayer.underlayer
        src_process, dst_process = cls.ip2pid(ip.src), cls.ip2pid(ip.dst)        
        message = {'DUMMY': 'DUMMY'}
        event = PacketEvent.from_message(src_process, dst_process, message)
        return event

    
class DummyPacket(Packet):
    name = 'DummyPacket'
    longname = 'DummyPacket'
    def do_dissect_payload(self, s):
       try:
           self.event = Util.make_event(self)
           self.add_payload(Raw(s))
       except Exception as e:
           LOG.exception(e)
           raise e

class DummyInspector(EtherInspectorBase):
    def __init__(self):
        super(DummyInspector, self).__init__(zmq_addr=ZMQ_ADDR)
        bind_layers(TCP, DummyPacket)
    
    def map_packet_to_event(self, pkt):
        """
        return None if this packet is NOT interesting at all.
        """
        if pkt.haslayer(DummyPacket):
            return pkt[DummyPacket].event
        else:
            # LOG.debug('%s unknown packet: %s', self.__class__.__name__, pkt.mysummary())                        
            return None


if __name__ == '__main__':
    d = DummyInspector()
    d.start()

#!/usr/bin/env python
import os

ZMQ_ADDR = os.getenv('NMZ_ETHER_ZMQ_ADDR')

from hexdump import hexdump
import pynmz
from pynmz.inspector.ether import EtherInspectorBase
from pynmz.signal.event import PacketEvent

LOG = pynmz.LOG.getChild(__name__)

class Zk2080Inspector(EtherInspectorBase):
    # @Override
    def map_packet_to_event(self, packet):
        src, dst = packet['IP'].src, packet['IP'].dst
        sport, dport = packet['TCP'].sport, packet['TCP'].dport
        payload = packet['TCP'].payload

        ## heuristic: FLE ports tend to be these ones. (PortAssignment.java)
        fle_ports = (11223, 11226, 11229, 11232)

        if (sport in fle_ports or dport in fle_ports) and payload:
            src_entity = 'entity-%s:%d' % (src, sport)
            dst_entity = 'entity-%s:%d' % (dst, dport)
            ## TODO: use zktraffic to parse the payload
            ## Currently zktraffic does not work well, because some packets get corked when the delay is injected.
            d = {'payload': hexdump(str(payload), result='return')}
            deferred_event = PacketEvent.from_message(src_entity, dst_entity, d)

            LOG.info('defer FLE packet: %s', deferred_event)

            return deferred_event
        else:
            return None


if __name__ == '__main__':
    d = Zk2080Inspector(zmq_addr=ZMQ_ADDR)
    d.start()

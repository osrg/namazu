#!/usr/bin/env python
import os

ZMQ_ADDR = os.getenv('NMZ_ETHER_ZMQ_ADDR')

import pynmz
from pynmz.inspector.zookeeper import ZkEtherInspector

LOG = pynmz.LOG.getChild(__name__)


class Zk2212Inspector(ZkEtherInspector):

    def __init__(self, zmq_addr):
        super(Zk2212Inspector, self).__init__(zmq_addr, ignore_pings=True)

    # @Override
    def map_zktraffic_message_to_entity_ids(self, zt_msg):
        src, dst = super(Zk2212Inspector, self).map_zktraffic_message_to_entity_ids(zt_msg)

        def try_map(s):
            ## FIXME: move these ones to the config file
            if s.startswith('entity-192.168.42.1:'):
                return 'zk1'
            elif s.startswith('entity-192.168.42.2:'):
                return 'zk2'
            elif s.startswith('entity-192.168.42.3:'):
                return 'zk3'
            return s

        new_src, new_dst = try_map(src), try_map(dst)
        LOG.debug('map_zktraffic_message_to_entity_ids: src=%s->%s, dst=%s->%s',
                  src, new_src, dst, new_dst)
        return new_src, new_dst

if __name__ == '__main__':
    d = Zk2212Inspector(zmq_addr=ZMQ_ADDR)
    d.start()

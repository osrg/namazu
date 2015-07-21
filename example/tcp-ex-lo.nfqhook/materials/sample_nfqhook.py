#!/usr/bin/env python
import os

NFQ_NUMBER=42
ZMQ_ADDR=os.getenv('EQ_ETHER_ZMQ_ADDR')

import pyearthquake
from pyearthquake.middlebox.nfqhook import NFQHook
LOG = pyearthquake.LOG.getChild(__name__)

if __name__ == '__main__':
    LOG.info("Please run `iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num %d` before running this hook.", NFQ_NUMBER)
    hook = NFQHook(nfq_number=NFQ_NUMBER, zmq_addr=ZMQ_ADDR)
    hook.start()

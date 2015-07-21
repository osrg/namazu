#!/usr/bin/env python

## FIXME: move these ones to the config file
NFQ_NUMBER=42
ZMQ_ADDR='ipc:///tmp/eq/ether_inspector'

import pyearthquake
from pyearthquake.middlebox.nfqhook import NFQHook
LOG = pyearthquake.LOG.getChild(__name__)

if __name__ == '__main__':
    LOG.info("Please run `iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num 42` before running this hook.")
    LOG.info("and then run `sudo -u nfqhooked SOME_COMMAND`")
    hook = NFQHook(nfq_number=NFQ_NUMBER, zmq_addr=ZMQ_ADDR)
    hook.start()

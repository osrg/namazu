#!/usr/bin/env python
import os
ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')
## FIXME: move to the config file
NFQ_NUMBER=42

from pyearthquake.middlebox.nfqhook import NFQHook

if __name__ == '__main__':
    hook = NFQHook(nfq_number=NFQ_NUMBER, zmq_addr=ZMQ_ADDR)
    hook.start()

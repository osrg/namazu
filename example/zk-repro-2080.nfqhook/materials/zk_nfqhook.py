#!/usr/bin/env python
import os
ZMQ_ADDR = os.getenv('EQ_ETHER_ZMQ_ADDR')
NFQ_NUMBER = int(os.getenv('EQ_NFQ_NUMBER'))

from pyearthquake.middlebox.nfqhook import NFQHook

if __name__ == '__main__':
    hook = NFQHook(nfq_number=NFQ_NUMBER, zmq_addr=ZMQ_ADDR)
    hook.start()

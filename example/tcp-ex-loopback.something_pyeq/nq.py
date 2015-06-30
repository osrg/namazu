#!/usr/bin/env python
Q_NUM=42
from netfilterqueue import NetfilterQueue
from scapy.all import *
import hexdump as hd


def print_and_accept(pkt):
    print('========== IP Packet ==========')
    ## NOTE: there is no ethernet header
    ip_str=pkt.get_payload()
    hd.hexdump(ip_str)    
    ip=IP(ip_str)
    print ip.summary()
    pkt.accept()

nfqueue = NetfilterQueue()
nfqueue.bind(Q_NUM, print_and_accept)
try:
    nfqueue.run()
except KeyboardInterrupt:
    print
    

#!/usr/bin/env python

## FIXME: move these ones to the config file
NFQ_NUMBER=42
SLEEP_SECS=0.03
SOCK_BUF_SIZE=65536

import pyearthquake
from pyearthquake.middlebox.nfq import NFQ
LOG = pyearthquake.LOG.getChild(__name__)
import socket
from scapy.all import IP
from time import sleep

def cb(qh, nfmsg, nfad, data):
    """
    int nfq_callback(struct nfq_q_handle *qh,
                     struct nfgenmsg *nfmsg,
                     struct nfq_data *nfad, void *data);
    """
    payload = NFQ.cb_get_payload(nfad)
    packet_id = NFQ.cb_get_packet_id(nfad)
    ip = IP(payload)
    LOG.info("ID %d: %s", packet_id, ip.summary())
    sleep(SLEEP_SECS)
    NFQ.cb_set_verdict(qh, packet_id, NFQ.NF_ACCEPT)
    return 1

cb_c = NFQ.CALLBACK_CFUNCTYPE(cb) # https://github.com/JohannesBuchner/PyMultiNest/issues/5


if __name__ == '__main__':
    LOG.info("Please run `iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num 42` before running this hook.")
    LOG.info("and then run `sudo -u nfqhooked SOME_COMMAND`")

    nfq = NFQ(NFQ_NUMBER, cb_c)
    s = socket.fromfd(nfq.fd, socket.AF_UNIX, socket.SOCK_STREAM)
    while True:
        d = s.recv(SOCK_BUF_SIZE)
        assert d
        nfq.handle_packet(d)
    s.close()
    nfq.close()

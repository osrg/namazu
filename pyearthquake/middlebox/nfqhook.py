import ctypes
import sys

import eventlet

from eventlet.green import socket

from .. import LOG as _LOG
from pyearthquake.middlebox.internal.nfq import NFQ
from pyearthquake.middlebox.internal.zmqclient import ZMQClientBase

LOG = _LOG.getChild(__name__)


def nfq_hook_cb(qh, nfmsg, nfad, data):
    """
    int nfq_callback(struct nfq_q_handle *qh,
                     struct nfgenmsg *nfmsg,
                     struct nfq_data *nfad, void *data);
    """
    packet_id = NFQ.cb_get_packet_id(nfad)
    ip_bytes = NFQ.cb_get_payload(nfad)
    this_id = data
    this = ctypes.cast(this_id, ctypes.py_object).value
    assert isinstance(this, NFQHook)
    this.send_ip_packet_to_inspector(packet_id, ip_bytes)
    return 1


nfq_hook_cb_c = NFQ.CALLBACK_CFUNCTYPE(nfq_hook_cb)  # https://github.com/JohannesBuchner/PyMultiNest/issues/5


class NFQHookZMQC(ZMQClientBase):
    def __init__(self, zmq_addr, nfq):
        super(NFQHookZMQC, self).__init__(zmq_addr)
        self.nfq = nfq

    def on_accept(self, packet_id, eth_bytes, metadata=None):
        NFQ.cb_set_verdict(self.nfq.qh, packet_id, NFQ.NF_ACCEPT)

    def on_drop(self, packet_id, eth_bytes, metadata=None):
        NFQ.cb_set_verdict(self.nfq.qh, packet_id, NFQ.NF_DROP)


class NFQHook(object):
    NFQ_SOCKET_BUFFER_SIZE = 1024 * 4

    def __init__(self, nfq_number, zmq_addr):
        LOG.info('NFQ Number: %s', nfq_number)
        self.self_id = id(self)
        self.nfq = NFQ(nfq_number,
                       nfq_hook_cb_c,
                       ctypes.c_void_p(self.self_id))
        self.zmq_client = NFQHookZMQC(zmq_addr, self.nfq)

    def start(self):
        zmq_worker_handle = self.zmq_client.start()
        nfq_worker_handle = eventlet.spawn(self._nfq_worker)
        nfq_worker_handle.wait()
        zmq_worker_handle.wait()
        raise RuntimeError('should not reach here')

    def _nfq_worker(self):
        s = socket.fromfd(self.nfq.fd, socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            while True:
                LOG.debug('NFQ Recv')
                nfad = s.recv(self.NFQ_SOCKET_BUFFER_SIZE)
                self.nfq.handle_packet(nfad)
            LOG.error('NFQ Worker loop leave!')
        finally:
            LOG.error('NFQ Worker closing socket!')
            s.close()
            self.nfq.close()
            sys.exit(1)

    def send_ip_packet_to_inspector(self, packet_id, ip_bytes):
        LOG.debug('Sending packet %d to inspector', packet_id)

        # from hexdump import hexdump
        # for l in hexdump(ip_bytes, result='generator'):
        #     LOG.debug(l)

        # TODO: eliminate dummy eth header
        dummy_eth = '\xff\xff\xff\xff\xff\xff' + '\x00\x00\x00\x00\x00\x00' + '\x08\x00' + ip_bytes
        self.zmq_client.send(packet_id, dummy_eth)

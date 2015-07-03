import eventlet
from eventlet.green import zmq, socket
from scapy.all import Ether, IP
from hexdump import hexdump

from .. import LOG as _LOG
from .nfq import NFQ

LOG = _LOG.getChild(__name__)


class NFQHook(object):
    NFQ_SOCKET_BUFFER_SIZE = 65536
    
    def __init__(self, nfq_number, zmq_addr):
        self.nfq_number = nfq_number
        self.zmq_addr = zmq_addr

    def start(self):
        zmq_worker_handle = self.start_zmq_worker()
        nfq_worker_handle = self.start_nfq_worker()
        zmq_worker_handle.wait()
        raise RuntimeError('should not reach here')
        
    def start_zmq_worker(self):
        LOG.info('Inspector ZMQ Address: %s', self.zmq_addr)
        self.zmq_ctx = zmq.Context()
        self.zs = self.zmq_ctx.socket(zmq.PAIR)
        self.zs.connect(self.zmq_addr)
        worker_handle = eventlet.spawn(self._zmq_worker)
        return worker_handle
        
    def start_nfq_worker(self):
        worker_handle = eventlet.spawn(self._nfq_worker)
        return worker_handle

    def _zmq_worker(self):
        while True:
            eth_frame = self.zs.recv()
            eth = Ether(eth_frame)
            assert eth.haslayer(IP)
            ip = eth[IP]
            LOG.debug('ZMQ IP frame: %s', ip.summary())
            
    def _nfq_worker(self):
        LOG.info('NFQ Number: %s', self.nfq_number)
        cb = self.get_nfq_cb()        
        nfq = NFQ(self.nfq_number, cb)        
        s = socket.fromfd(nfq.fd, socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            while True:
                nfad = s.recv(self.NFQ_SOCKET_BUFFER_SIZE)
                nfq.handle_packet(nfad)
            LOG.error('NFQ Worker loop leave!')
        finally:
            LOG.error('NFQ Worker closing socket!')            
            s.close()
            nfq.close()

    def get_nfq_cb(self):
        def cb(qh, nfmsg, nfad, data):
            """
            int nfq_callback(struct nfq_q_handle *qh,
                             struct nfgenmsg *nfmsg,
                             struct nfq_data *nfad, void *data);
            """
            LOG.debug("----- NFQ CB enter(self.nfq_number=%d) -----", self.nfq_number)
            ip_frame = NFQ.cb_get_payload(nfad)
            LOG.debug("NFQ CB ip frame got")            
            packet_id = NFQ.cb_get_packet_id(nfad)
            LOG.debug("NFQ CB ip packet id got")                        
            ip = IP(ip_frame)
            LOG.debug("NFQ IP frame: %s (ID=%d)" % (ip.summary(), packet_id))
            hexdump(ip_frame)
            NFQ.cb_set_verdict(qh, packet_id, NFQ.NF_ACCEPT)
            LOG.debug("----- NFQ CB leave -----")            
            return 1
        
        return NFQ.CALLBACK_CFUNCTYPE(cb)

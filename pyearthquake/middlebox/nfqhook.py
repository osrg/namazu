import ctypes
import eventlet
from eventlet.green import zmq, socket
from scapy.all import Ether, IP
from hexdump import hexdump

from .. import LOG as _LOG
from .nfq import NFQ

LOG = _LOG.getChild(__name__)

def nfq_hook_cb(qh, nfmsg, nfad, data):
    """
    int nfq_callback(struct nfq_q_handle *qh,
                     struct nfgenmsg *nfmsg,
                     struct nfq_data *nfad, void *data);
    """
    packet_id = NFQ.cb_get_packet_id(nfad)            
    LOG.debug('packet_id=%d', packet_id)
    ip_frame = NFQ.cb_get_payload(nfad)
    LOG.debug('ip_frame=len:%d', len(ip_frame))
    LOG.debug('Getting this')
    this_id = data
    LOG.debug('this_id=%s', this_id)
    this = ctypes.cast(this_id, ctypes.py_object).value
    LOG.debug('Got this')    
    assert isinstance(this, NFQHook)
    LOG.debug('sending packet %d', packet_id)    
    this.send_ip_frame_to_inspector(packet_id, ip_frame)
    LOG.debug('sent packet %d', packet_id)            
    return 1


class NFQHook(object):
    NFQ_SOCKET_BUFFER_SIZE = 65536
    
    def __init__(self, nfq_number, zmq_addr):
        self.nfq_number = nfq_number
        self.zmq_addr = zmq_addr
        self.pendings={} # key: hash of ip frame, value: list of packet id

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
        LOG.info('NFQ Number: %s', self.nfq_number)
        self_id = id(self)
        LOG.debug('NFQ Worker self id: %s', self_id)
        self.nfq = NFQ(self.nfq_number, 
                       NFQ.CALLBACK_CFUNCTYPE(nfq_hook_cb), 
                       ctypes.c_void_p(self_id))
        worker_handle = eventlet.spawn(self._nfq_worker)
        return worker_handle

    def _zmq_worker(self):
        while True:
            dummy_eth_frame = self.zs.recv() #TODO: recv packet id and accept/drop
            dummy_eth = Ether(dummy_eth_frame)
            assert dummy_eth.haslayer(IP)
            ip = dummy_eth[IP]
            ip_frame_hash = hash(str(ip))
            if not ip_frame_hash in self.pendings:
                LOG.error('Received Unknown IP frame: %s', ip.summary())
                continue

            for packet_id in self.pendings[ip_frame_hash]:
                LOG.debug('Accepting IP frame: %s (packet id=%d)', ip.summary(), packet_id)                
                NFQ.cb_set_verdict(self.nfq.qh, packet_id, NFQ.NF_ACCEPT)
            del self.pendings[ip_frame_hash]                
            
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

    def send_ip_frame_to_inspector(self, packet_id, ip_frame):
        LOG.debug("Sending NFQ IP frame: (ID=%d)" % (packet_id))
        ip_frame_hash = hash(ip_frame)
        ip = IP(ip_frame)
        LOG.debug("Sending NFQ IP frame: %s (ID=%d, hash=%d)" % (ip.summary(), packet_id, ip_frame_hash))        
        if not ip_frame_hash in self.pendings:
            self.pendings[ip_frame_hash] = []
        self.pendings[ip_frame_hash].append(packet_id) # multiple packet id can have same frame
        dummy_eth = Ether()/ip
        self.zs.send(str(dummy_eth)) #TODO: send packet id
        LOG.debug("Sent NFQ IP frame: (ID=%d)" % (packet_id))        
        

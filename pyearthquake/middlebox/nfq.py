#!/usr/bin/python

from socket import AF_INET, ntohl
from ctypes import *
import logging
import prctl

from .. import LOG as _LOG
LOG = _LOG.getChild(__name__)
#LOG.setLevel(logging.INFO)

class NFQ(object):
    """
    low-level NFQ
    TODO: check SegV
    """
    NF_DROP = 0
    NF_ACCEPT = 1
    NF_STOLEN = 2
    NF_QUEUE = 3
    NF_REPEAT = 4
    NF_STOP = 5  
  
    NFQNL_COPY_NONE = 0
    NFQNL_COPY_META = 1
    NFQNL_COPY_PACKET = 2

    class NFQNL_MSG_PACKET_HDR(Structure):
        _fields_ = [("packet_id", c_uint32),
                    ("hw_protocol", c_uint16),
                    ("hook", c_uint8)]

    dll = CDLL('libnetfilter_queue.so.1')
    dll.nfq_get_msg_packet_hdr.argtypes = [c_void_p]
    dll.nfq_get_msg_packet_hdr.restype = POINTER(NFQNL_MSG_PACKET_HDR)
    dll.nfq_get_payload.argtypes = [c_void_p, POINTER(POINTER(c_char))]
    dll.nfq_get_payload.restype = c_int

    CALLBACK_CFUNCTYPE = CFUNCTYPE(c_int,
                                   c_void_p, #qh
                                   c_void_p, #nfmsg
                                   c_void_p, #nfad
                                   c_void_p) #data

    def __init__(self, q_num, cb, cb_data=c_void_p(None)):
        self._check_cap()
        self.h = self._make_handle()
        self.qh = self._make_queue_handle(self.h, q_num, cb, cb_data)
        self.fd = self.dll.nfq_fd(self.h)

    @classmethod
    def _check_cap(cls):
        assert prctl.cap_permitted.net_admin, "missing CAP_NET_ADMIN"
        assert prctl.cap_permitted.net_raw, "missing CAP_NET_RAW"
        
        
    @classmethod
    def _make_handle(cls):
        LOG.debug('Calling nfq_open()')
        h = cls.dll.nfq_open()
        LOG.debug('Called nfq_open=%d', h)
        assert h, "h=%d" % h
        LOG.debug('Calling nfq_unbind_pf(%d, AF_INET)', h)
        unbound = cls.dll.nfq_unbind_pf(h, AF_INET)
        LOG.debug('Called nfq_unbind_pf=%d', unbound)
        assert unbound >= 0, "unbound=%d" % unbound
        LOG.debug('Calling nfq_bind_pf(%d, AF_INET)', h)
        bound = cls.dll.nfq_bind_pf(h, AF_INET)
        LOG.debug('Called nfq_bind_pf=%d', bound)
        assert bound >= 0, "bound=%d" % bound
        return h

    @classmethod
    def _make_queue_handle(cls, h, q_num, cb, data):
        LOG.debug('Calling nfq_create_queue(%s, %s, %s, %s)', h, q_num, cb, data)        
        qh = cls.dll.nfq_create_queue(h, q_num, cb, data)
        LOG.debug('Called nfq_create_queue=%d', qh)                
        assert qh, "qh=%d" % qh
        LOG.debug('Calling nfq_set_mode(%d, NFQNL_COPY_PACKET)', qh)                
        mode_set = cls.dll.nfq_set_mode(qh, cls.NFQNL_COPY_PACKET)
        LOG.debug('Called nfq_set_mode=%d', mode_set)
        assert mode_set >= 0, "mode_set=%d" % mode_set
        return qh

    def close(self):
        LOG.debug('Calling nfq_close(%d)', self.h)                                
        self.dll.nfq_close(self.h)
        LOG.debug('Called nfq_close()')                                        

    def handle_packet(self, buf):
        assert self.h
        assert buf
        LOG.debug('Calling nfq_handle_packet(%s, buf, %s)', self.h, len(buf))
        self.dll.nfq_handle_packet(self.h, buf, len(buf))
        LOG.debug('Called nfq_handle_packet')                        

    @classmethod
    def cb_get_payload(cls, nfad):
        """
        int nfq_get_payload(struct nfq_data * nfad, unsigned char ** data)
        """
        assert nfad
        payload = POINTER(c_char)()
        LOG.debug('Calling nfq_get_payload')        
        len = cls.dll.nfq_get_payload(nfad, byref(payload))
        LOG.debug('Called nfq_get_payload')
        ret_str = string_at(payload, len)
        return ret_str

    @classmethod
    def cb_get_packet_id(cls, nfad):
        assert nfad
        LOG.debug('Calling nfq_get_msg_packet_hdr')
        phdr = cls.dll.nfq_get_msg_packet_hdr(nfad)
        LOG.debug('Called nfq_get_msg_packet_hdr')
        assert phdr
        packet_id = ntohl(phdr.contents.packet_id)
        return packet_id

    @classmethod
    def cb_set_verdict(cls, qh, packet_id, verdict):
        """
        int nfq_set_verdict (struct nfq_q_handle * qh,
                             u_int32_t id,
                             u_int32_t verdict,
                             u_int32_t data_len,
                             const unsigned char * buf)
        """
        LOG.debug('Calling nfq_set_verdict')
        ret = cls.dll.nfq_set_verdict(qh, packet_id, verdict, 0, 0)
        LOG.debug('Called nfq_set_verdict')
        return ret


if __name__ == "__main__":
    Q_NUM=42
    SOCK_BUF_SIZE=65536
    import socket
    from scapy.all import IP
    from hexdump import hexdump

    def cb(qh, nfmsg, nfad, data):
        """
        int nfq_callback(struct nfq_q_handle *qh,
                         struct nfgenmsg *nfmsg,
                         struct nfq_data *nfad, void *data);
        """
        print "===CB==="
        LOG.info("CB called with data=%s", data)
        payload = NFQ.cb_get_payload(nfad)
        packet_id = NFQ.cb_get_packet_id(nfad)
        hexdump(payload)
        ip = IP(payload)
        LOG.info("ID %d: %s", packet_id, ip.summary())
        NFQ.cb_set_verdict(qh, packet_id, NFQ.NF_ACCEPT)
        return 1

    cb_c = NFQ.CALLBACK_CFUNCTYPE(cb) # https://github.com/JohannesBuchner/PyMultiNest/issues/5
    nfq = NFQ(Q_NUM, cb_c)
    s = socket.fromfd(nfq.fd, socket.AF_UNIX, socket.SOCK_STREAM)
    while True:
            d = s.recv(SOCK_BUF_SIZE)
            assert d
            nfq.handle_packet(d)
    s.close()
    nfq.close()

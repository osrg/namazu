#!/usr/bin/python

import socket
from ctypes import *
import prctl

class NFQ(object):
    """
    low-level NFQ
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

    dll = CDLL('libnetfilter_queue.so.1')
    callback_type = CFUNCTYPE(c_int,
                              c_void_p, #qh
                              c_void_p, #nfmsg
                              c_void_p, #nfad
                              c_void_p) #data

    def __init__(self, q_num, py_cb):
        self._check_cap()
        self.h = self._make_handle()
        cb = self.callback_type(py_cb)
        self.qh = self._make_queue_handle(self.h, q_num, cb)
        self.fd = self.dll.nfq_fd(self.h)

    @classmethod
    def _check_cap(cls):
        assert prctl.cap_permitted.net_admin, "missing CAP_NET_ADMIN"
        assert prctl.cap_permitted.net_raw, "missing CAP_NET_RAW"
        
    @classmethod
    def _make_handle(cls):
        h = cls.dll.nfq_open()
        assert h, "h=%d" % h
        unbound = cls.dll.nfq_unbind_pf(h, socket.AF_INET)
        assert unbound >= 0, "unbound=%d" % unbound
        bound = cls.dll.nfq_bind_pf(h, socket.AF_INET)
        assert bound >= 0, "bound=%d" % bound
        return h

    @classmethod
    def _make_queue_handle(cls, h, q_num, cb):
        qh = cls.dll.nfq_create_queue(h, q_num, cb, 0)
        assert qh, "qh=%d" % qh
        mode_set = cls.dll.nfq_set_mode(qh, cls.NFQNL_COPY_PACKET)
        assert mode_set >= 0, "mode_set=%d" % mode_set
        return qh

    def close(self):
        self.dll.nfq_close(self.h)

    def handle_packet(self, buf):
        self.dll.nfq_handle_packet(self.h, buf, len(buf))

    @classmethod
    def cb_get_payload(cls, nfad):
        """
        int nfq_get_payload(struct nfq_data * nfad, unsigned char ** data)
        """
        assert nfad
        payload = POINTER(c_char)()
        func = cls.dll.nfq_get_payload
        func.argtypes = [c_void_p, POINTER(POINTER(c_char))]
        func.restype = c_int
        len = func(nfad, byref(payload))
        ret_str = string_at(payload, len)
        return ret_str

    @classmethod
    def cb_get_packet_id(cls, nfad):
        assert nfad
        class NFQNL_MSG_PACKET_HDR(Structure):
            _fields_ = [("packet_id", c_uint32),
                        ("hw_protocol", c_uint16),
                        ("hook", c_uint8)]
        func = cls.dll.nfq_get_msg_packet_hdr
        func.argtypes=[c_void_p]
        func.restype = POINTER(NFQNL_MSG_PACKET_HDR)
        phdr = func(nfad)
        assert phdr
        packet_id = socket.ntohl(phdr.contents.packet_id)
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
        return cls.dll.nfq_set_verdict(qh, packet_id, verdict, 0, 0)


if __name__ == "__main__":
    Q_NUM=42
    SOCK_BUF_SIZE=65536
    from scapy.all import IP
    from hexdump import hexdump

    def cb(qh, nfmsg, nfad, data):
        """
        int nfq_callback(struct nfq_q_handle *qh,
                         struct nfgenmsg *nfmsg,
                         struct nfq_data *nfad, void *data);
        """
        payload = NFQ.cb_get_payload(nfad)
        packet_id = NFQ.cb_get_packet_id(nfad)
        hexdump(payload)
        ip = IP(payload)
        print "ID %d: %s" % (packet_id, ip.summary())
        NFQ.cb_set_verdict(qh, packet_id, NFQ.NF_ACCEPT)
        return 1

    nfq = NFQ(Q_NUM, cb)
    s = socket.fromfd(nfq.fd, socket.AF_UNIX, socket.SOCK_STREAM)
    while True:
            d = s.recv(SOCK_BUF_SIZE)
            assert d
            nfq.handle_packet(d)
    s.close()
    nfq.close()

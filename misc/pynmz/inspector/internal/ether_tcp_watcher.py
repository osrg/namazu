import ryu.lib.packet.packet

from pynmz import LOG as _LOG
LOG = _LOG.getChild(__name__)


class TCPWatcher(object):
    RST = 0x04

    def __init__(self):
        self._last_tcp_table = {}  # Map<Tuple<string, int, string, int>, TCP>

    def _inspect_ryu(self, eth_bytes):
        """
        ryu-based inspector ( this does NOT call EtherInspector.inspect())

        self._on_recv_with_tcp_handling() uses this function to avoid to call scapy inspector
        before detecting TCP retransmission, which may cause unexpected side-effects.

        Should be NOT called from any function other than self._on_recv_with_tcp_handling()
        """
        ryu_pkt = ryu.lib.packet.packet.Packet(eth_bytes)
        return ryu_pkt

    @classmethod
    def _is_tcp_retrans(cls, tcp, last_tcp):
        """
        tcp: ryu.lib.packet.tcp.tcp
        """
        return tcp and last_tcp and \
            tcp.seq == last_tcp.seq and \
            tcp.ack == last_tcp.ack and \
            tcp.bits == last_tcp.bits

    def _get_last_tcp(self, ip, tcp):
        """
        tcp: ryu.lib.packet.tcp.tcp
        """
        try:
            return self._last_tcp_table[ip.src, tcp.src_port, ip.dst, tcp.dst_port]
        except KeyError:
            return None

    def _set_last_tcp(self, ip, tcp):
        """
        tcp: ryu.lib.packet.tcp.tcp
        """
        self._last_tcp_table[ip.src, tcp.src_port, ip.dst, tcp.dst_port] = tcp

    def _del_last_tcp(self, ip, tcp):
        """
        tcp: ryu.lib.packet.tcp.tcp
        """
        try:
            del self._last_tcp_table[
                ip.src, tcp.src_port, ip.dst, tcp.dst_port]
        except KeyError:
            return

    def dumb_handler(metadata, eth_bytes):
        LOG.warn('dumb_handler called!!')

    def on_recv(self, metadata, eth_bytes, default_handler=dumb_handler, retrans_handler=dumb_handler):
        # Do NOT call scapy inspector before detecting TCP retransmission
        ryu_pkt = self._inspect_ryu(eth_bytes)
        ip = ryu_pkt.get_protocol(ryu.lib.packet.ipv4.ipv4)
        tcp = ryu_pkt.get_protocol(ryu.lib.packet.tcp.tcp)
        if tcp:
            last_tcp = self._get_last_tcp(ip, tcp)
            if self._is_tcp_retrans(tcp, last_tcp):
                LOG.debug('TCP retrans %s', tcp)
                retrans_handler(metadata, eth_bytes)
            else:  # not TCP retrans
                # LOG.debug('TCP NOT retrans %s [last=%s]', tcp, last_tcp)
                if tcp.bits & self.RST:
                    # if not last_tcp: LOG.debug('RST for unknown connection')
                    self._del_last_tcp(ip, tcp)
                else:  # not RST or no last TCP has been recorded
                    self._set_last_tcp(ip, tcp)
                default_handler(metadata, eth_bytes)
        else:  # not TCP
            default_handler(metadata, eth_bytes)

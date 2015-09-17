#!/usr/bin/env python

## FIXME: move these ones to the config file
ZMQ_ADDR='ipc:///tmp/eq/ether_inspector'

import colorama
from scapy.all import *
import pyearthquake
from pyearthquake.inspector.ether import *
from pyearthquake.signal.signal import *
from pyearthquake.signal.event import *
from pyearthquake.signal.action import *
import hexdump as hd # hexdump conflict with scapy.all.hexdump

import zktraffic # tested with b3e9dd0 (Jun 4 2015)
import zktraffic.base.client_message
import zktraffic.base.server_message
import zktraffic.zab.quorum_packet
import zktraffic.fle.message


LOG = pyearthquake.LOG.getChild(__name__)

# terrible table (TODO: move to config.json)
pid_table = {}

class Util(object):
    ## scapy hates inheritance from scapy.all.Packet class

    @classmethod
    def ip2pid(cls, ip):
        # TODO: move to pyearthquake.util
        try:
            return pid_table[ip]
        except KeyError:
            return '_unknown_process_%s' % (ip)


    @classmethod
    def make_zktraffic_packet(cls, klazz, s, zk_klazz):
        tcp, ip = klazz.underlayer, klazz.underlayer.underlayer
        zt = zk_klazz.from_payload(s, '%s:%d' % (ip.src, tcp.sport), '%s:%d' % (ip.dst, tcp.dport), time.time())
        return zt

    @classmethod
    def _make_message_from_zt(cls, klazz, zt):
        msg = { 'class_group': klazz.name, 'class': zt.__class__.__name__ }
        ignore_keys = ('timestr', 'src', 'dst', 'length', 'session_id', 'client_id', 'txn_time', 'txn_zxid', 'timeout', 'timestamp', 'ip', 'port', 'session', 'client') # because client port may differ
        def gen():
            for k in dir(zt):
                v = getattr(zt, k)
                cond = (isinstance(v, int) or isinstance(v, basestring)) and \
                       not k.isupper() and not k.startswith('_') and not '_literal' in k \
                       and not k == 'type'
                if cond:
                    alt_k = '%s_literal' % k
                    if hasattr(zt, alt_k):
                        v = getattr(zt, alt_k)
                    yield k, v
        for k, v in gen():
            if k in ignore_keys: continue
            if k == 'zxid':
                msg['zxid_hi'] = v >> 32
                msg['zxid_low'] = v & 0xFFFF
                continue
            msg[k] = v
        return msg

    @classmethod
    def make_zktraffic_event(cls, klazz, zt):
        tcp, ip = klazz.underlayer, klazz.underlayer.underlayer
        src_process, dst_process = cls.ip2pid(ip.src), cls.ip2pid(ip.dst)
        message = cls._make_message_from_zt(klazz, zt)
        event = PacketEvent.from_message(src_process, dst_process, message)
        return event

    FOUR_LETTER_WORDS = ('dump', 'envi', 'kill', 'reqs', 'ruok', 'srst', 'stat')
    @classmethod
    def make_four_letter_event(cls, klazz, four_letter, reply=None):
        tcp, ip = klazz.underlayer, klazz.underlayer.underlayer
        src_process, dst_process = cls.ip2pid(ip.src), cls.ip2pid(ip.dst)
        message = {'class_group': 'ZkFourLetterPacket', 'class': four_letter}
        if reply: message['reply'] = reply
        event = PacketEvent.from_message(src_process, dst_process, message)
        return event


    @classmethod
    def make_summary(cls, klazz):
        s = str(klazz.event)
        return s

    @classmethod
    def print_dissection_error(cls, inst, s, e):
        tcp, ip = inst.underlayer, inst.underlayer.underlayer
        src, dst = intern("%s:%d" % (ip.src, tcp.sport)), intern("%s:%d" % (ip.dst, tcp.dport))
        LOG.error('Error while dissecting %s (%s->%s)', inst.name, src, dst)
        LOG.exception(e)
        try:
            LOG.error('Hexdump (%d bytes)', len(s))
            for line in hd.hexdump(s, result='generator'):
                LOG.error('%s', line)
        except:
            LOG.error('Error while hexdumping', exc_info=True)


class ZkQuorumPacket(Packet):
    name = 'ZkQuorumPacket'
    longname = 'ZooKeeper Zab Quorum Packet'
    def do_dissect_payload(self, s):
       try:
           self.zt = Util.make_zktraffic_packet(self, s, zktraffic.zab.quorum_packet.QuorumPacket)
           self.event = Util.make_zktraffic_event(self, self.zt)
           self.add_payload(Raw(s))
       except Exception as e:
           Util.print_dissection_error(self, s, e)
           raise e
            
    def mysummary(self):
        return Util.make_summary(self)
        

class ZkFLEPacket(Packet):
    name = 'ZkFLEPacket'
    longname = 'ZooKeeper Fast Leader Election Packet'
    def do_dissect_payload(self, s):
       try:
           self.zt = Util.make_zktraffic_packet(self, s, zktraffic.fle.message.Message)
           self.event = Util.make_zktraffic_event(self, self.zt)
           self.add_payload(Raw(s))
       except Exception as e:
           Util.print_dissection_error(self, s, e)
           raise e

    def mysummary(self):
        return Util.make_summary(self)


_requests_xids = defaultdict(dict)
_four_letter_mode = defaultdict(dict) # key: client addr, val: four letter
class ZkPacket(Packet):
    """
    port 2181 packet
    
    ATTENTION! this dissector has side-effects
    """
    name = 'ZkPacket'
    longname = 'ZooKeeper C/S Packet'
        
    def _get_four_letter_mode(self, client):
        global _four_letter_mode
        if client in _four_letter_mode:
            return _four_letter_mode[client]
        else:
            return None

    def _set_four_letter_mode(self, client, four_letter=None):
        global _four_letter_mode
        if four_letter:
            _four_letter_mode[client] = four_letter
        else:
            del _four_letter_mode[client]

    def do_dissect_payload(self, s):
        try:
            ## this complicated logic is based on zktraffic.base.sniffer.Sniffer
            global _requests_xids
            tcp, ip = self.underlayer, self.underlayer.underlayer
            src, dst = intern("%s:%d" % (ip.src, tcp.sport)), intern("%s:%d" % (ip.dst, tcp.dport))
            assert tcp.dport == 2181 or tcp.sport == 2181
            if tcp.dport == 2181:
                client, server = src, dst
                if s.startswith(Util.FOUR_LETTER_WORDS):
                    self._set_four_letter_mode(client, s[0:4])
                    self.event = Util.make_four_letter_event(self, s[0:4])
                else:
                    self.zt = zktraffic.base.client_message.ClientMessage.from_payload(s, client, server)
                    if not self.zt.is_ping and not self.zt.is_close:
                        _requests_xids[self.zt.client][self.zt.xid] = self.zt.opcode
                    self.event = Util.make_zktraffic_event(self, self.zt)
            if tcp.sport == 2181:
                client, server = dst, src
                four_letter = self._get_four_letter_mode(client)
                if four_letter:
                    self.event = Util.make_four_letter_event(self, four_letter, reply=s)
                    self._set_four_letter_mode(client, None)
                else:
                    requests_xids = _requests_xids.get(client, {})
                    self.zt = zktraffic.base.server_message.ServerMessage.from_payload(s, client, server, requests_xids)
                    self.event = Util.make_zktraffic_event(self, self.zt)
            self.add_payload(Raw(s))
        except Exception as e:
            Util.print_dissection_error(self, s, e)
            raise e
            
    def mysummary(self):
        return Util.make_summary(self)

IGNORE_PING = False

class ZkInspector(EtherInspectorBase):
    def __init__(self):
        super(ZkInspector, self).__init__(zmq_addr=ZMQ_ADDR)
        self.regist_layer_on_tcp(ZkQuorumPacket, 2888)
        self.regist_layer_on_tcp(ZkFLEPacket, 3888)
        self.regist_layer_on_tcp(ZkPacket, 2181)

    def _print_packet_as(self, pkt, klazz, color):
        LOG.debug(color + 'Received packet, event=%s' + colorama.Style.RESET_ALL,
                  pkt[klazz].mysummary())
    
    def _map_ZkQuorumPacket_to_event(self, pkt):
        event = pkt[ZkQuorumPacket].event        
        msg = event.option['message']
        if IGNORE_PING and (msg['class_group'] == 'ZkQuorumPacket' and msg['class'] == 'Ping'):
            return None
        self._print_packet_as(pkt, ZkQuorumPacket, colorama.Back.WHITE + colorama.Fore.BLACK)
        return event

    def _map_ZkFLEPacket_to_event(self, pkt):
        event = pkt[ZkFLEPacket].event        
        self._print_packet_as(pkt, ZkFLEPacket, colorama.Back.CYAN + colorama.Fore.BLACK)
        return event

    def _map_ZkPacket_to_event(self, pkt):
        event = pkt[ZkPacket].event
        msg = event.option['message']        
        self._print_packet_as(pkt, ZkPacket, colorama.Back.BLUE + colorama.Fore.WHITE)
        if msg['class_group'] == 'ZkFourLetterPacket':
            return None
        return event
    
    def map_packet_to_event(self, pkt):
        """
        return None if this packet is NOT interesting at all.
        """
        if pkt.haslayer(ZkQuorumPacket):
            return self._map_ZkQuorumPacket_to_event(pkt)
        elif pkt.haslayer(ZkFLEPacket):
            return self._map_ZkFLEPacket_to_event(pkt)
        elif pkt.haslayer(ZkPacket):
            return self._map_ZkPacket_to_event(pkt)            
        else:
            # LOG.debug('%s unknown packet: %s', self.__class__.__name__, pkt.mysummary())                        
            return None


if __name__ == '__main__':
    d = ZkInspector()
    d.start()

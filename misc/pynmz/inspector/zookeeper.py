#!/usr/bin/env python
import colorama
import struct

import scapy.all
from zktraffic.network.sniffer import Sniffer
from zktraffic.base.client_message import ClientMessage
from zktraffic.base.server_message import ServerMessage
import zktraffic.fle.message as FLE
import zktraffic.zab.quorum_packet as ZAB
from zktraffic.base.network import BadPacket
from zktraffic.base.sniffer import Sniffer as ZKSniffer, SnifferConfig as ZKSnifferConfig
from zktraffic.omni.omni_sniffer import OmniSniffer

from .. import LOG as _LOG
from .ether import EtherInspectorBase
from ..signal.signal import ActionBase, DEFAULT_ORCHESTRATOR_URL
from ..signal.event import PacketEvent

LOG = _LOG.getChild(__name__)


class ZkEtherInspector(EtherInspectorBase):

    def __init__(self, zmq_addr, orchestrator_rest_url=DEFAULT_ORCHESTRATOR_URL,
                 entity_id='_namazu_ether_inspector',
                 ignore_pings=True, dump_bad_packet=False):
        super(ZkEtherInspector, self).__init__(zmq_addr, orchestrator_rest_url, entity_id)
        self.ignore_pings = ignore_pings
        self.dump_bad_packet = dump_bad_packet
        self._init_sniffer()

    def _init_sniffer(self):
        def fle_sniffer_factory(port):
            return Sniffer('dummy', port, FLE.Message, None, dump_bad_packet=False, start=False)

        def zab_sniffer_factory(port):
            return Sniffer('dummy', port, ZAB.QuorumPacket, None, dump_bad_packet=False, start=False)

        def zk_sniffer_factory(port):
            config = ZKSnifferConfig('dummy')
            config.track_replies = True
            config.zookeeper_port = port
            config.client_port = 0
            return ZKSniffer(config, None, None, None, error_to_stderr=True)

        self.sniffer = OmniSniffer(
            fle_sniffer_factory,
            zab_sniffer_factory,
            zk_sniffer_factory,
            dump_bad_packet=False,
                # ignored ,as we don't use sniffer.handle_packet()
            start=False)

    def map_zktraffic_message_to_entity_ids(self, zt_msg):
        """
        you should override this, if possible
        """
        src_entity = 'entity-%s' % zt_msg.src
        dst_entity = 'entity-%s' % zt_msg.dst
        return (src_entity, dst_entity)

    def map_zktraffic_message_to_dict(self, zt_msg):
        class_group = '_unknown'
        if isinstance(zt_msg, FLE.Message):
            class_group = 'FLE'
        elif isinstance(zt_msg, ZAB.QuorumPacket):
            class_group = 'ZAB'
        elif isinstance(zt_msg, ClientMessage):
            class_group = 'ClientMessage'
        elif isinstance(zt_msg, ServerMessage):
            class_group = 'ServerMessage'
        d = {'class_group': class_group, 'class': zt_msg.__class__.__name__}
        ignored_keys = (
            'type', 'timestr', 'src', 'dst', 'length', 'session_id', 'client_id', 'txn_time', 'txn_zxid', 'timeout',
            'timestamp', 'ip', 'port', 'session',
            'client',  # because client port may differ
            'passwd',  # may include non-ascii chars
        )

        def gen():
            for k in dir(zt_msg):
                v = getattr(zt_msg, k)
                cond1 = isinstance(v, int) or isinstance(
                    v, basestring)  # int or string
                cond2 = not k.isupper() and not k.startswith(
                    '_')  # not something like "FOO" or "_foo"
                cond3 = not '_literal' in k  # not something like "foo_literal"
                cond4 = not k in ignored_keys
                cond = cond1 and cond2 and cond3 and cond4
                if cond:
                    alt_k = '%s_literal' % k  # use "foo_literal" instead of "foo", if exists
                    if hasattr(zt_msg, alt_k):
                        v = getattr(zt_msg, alt_k)
                    yield k, v

        for k, v in gen():
            if k == 'zxid':
                if isinstance(v, str) and v.startswith('0x'):
                    v = int(v, 16)
                d['zxid_hi'] = v >> 32
                d['zxid_low'] = v & 0xFFFF
                continue
            d[k] = v

        return d

    def map_zktraffic_message_to_event(self, zt_msg):
        src_entity, dst_entity = self.map_zktraffic_message_to_entity_ids(
            zt_msg)
        d = self.map_zktraffic_message_to_dict(zt_msg)
        ## replay_hint is an optional string that helps replaying
        replay_hint = str(hash(frozenset(d.items())))
        event = PacketEvent.from_message(src_entity, dst_entity, d, replay_hint)

        if isinstance(zt_msg, FLE.Message):
            LOG.debug(colorama.Back.CYAN + colorama.Fore.BLACK +
                      'FLE: %s' + colorama.Style.RESET_ALL, event)
        elif isinstance(zt_msg, ZAB.QuorumPacket):
            LOG.debug(colorama.Back.WHITE + colorama.Fore.BLACK +
                      'ZAB: %s' + colorama.Style.RESET_ALL, event)
        elif isinstance(zt_msg, ClientMessage):
            LOG.debug(colorama.Back.BLUE + colorama.Fore.WHITE +
                      'CM: %s' + colorama.Style.RESET_ALL, event)
        elif isinstance(zt_msg, ServerMessage):
            LOG.debug(colorama.Back.RED + colorama.Fore.WHITE +
                      'SM: %s' + colorama.Style.RESET_ALL, event)
        else:
            LOG.debug('Unknown event %s', event)

        return event

    # @Override
    def map_packet_to_event(self, packet):
        """
        return None if this packet is NOT interesting at all.
        """
        try:
            raw_packet = scapy.all.Raw(str(packet))
            # NOTE: zktraffic expects this raw_packet rather than packet
            zt_msg = self.sniffer.message_from_packet(raw_packet)
            zt_msg.src = '%s:%d' % (
                packet[scapy.all.IP].src, packet[scapy.all.TCP].sport)
            zt_msg.dst = '%s:%d' % (
                packet[scapy.all.IP].dst, packet[scapy.all.TCP].dport)
            if self.ignore_pings:
                is_zab_ping = isinstance(zt_msg, ZAB.Ping)
                is_client_ping = isinstance(
                    zt_msg, ClientMessage) and zt_msg.is_ping
                is_server_ping = isinstance(
                    zt_msg, ServerMessage) and zt_msg.is_ping
                if is_zab_ping or is_client_ping or is_server_ping:
                    return None
            event = self.map_zktraffic_message_to_event(zt_msg)
            return event
        except (BadPacket, struct.error) as ex:
            # NOTE: ex happens on TCP SYN, RST and so on

            if len(ex.args) > 0:
                if 'Four letter request' in ex.args[0]:
                    return PacketEvent.from_message('_unknown', '_unknown', {'class_group': 'FourLetter', 'class': 'FourLetterRequest', 'data': packet.load})
                elif 'Four letter response' in ex.args[0]:
                    return PacketEvent.from_message('_unknown', '_unknown', {'class_group': 'FourLetter', 'class': 'FourLetterResponse', 'data': packet.load})

            if self.dump_bad_packet:
                raise ex  # the upper caller should print this
            return None

from abc import ABCMeta, abstractmethod
import six
import eventlet
from ryu.base import app_manager
from ryu.controller import ofp_event
from ryu.controller.handler import CONFIG_DISPATCHER, MAIN_DISPATCHER
from ryu.controller.handler import set_ev_cls
from ryu.ofproto import ofproto_v1_3, ofproto_v1_3_parser
from ryu.lib.packet import packet
from ryu.lib.packet import ethernet
from ryu.app.simple_switch_13 import SimpleSwitch13
from eventlet.green import zmq

DISABLE_INJECTION=0

@six.add_metaclass(ABCMeta)
class RyuOF13SwitchBase(SimpleSwitch13):
    """
    Inherits ryu.app.SimpleSwitch13.
    Tested with ryu 3.20.2 + OVS 2.3.1.
    """
    OFP = ofproto_v1_3
    OFP_PARSER = ofproto_v1_3_parser
    FLOW_PRIORITY_BASE = 10240
    
    @abstractmethod
    def __init__(self, matches, inspector_zmq_addr,  *args, **kwargs):
        super(RyuOF13SwitchBase, self).__init__(*args, **kwargs)
        raise NotImplementedError('Under ZMQ refactoring')
        self.matches = matches
        ### Start ZMQ Worker
        self.zmq_ctx = zmq.Context()
        self.zs = self.zmq_ctx.socket(zmq.PAIR)
        self.logger.info('Inspector ZMQ Address: %s', inspector_zmq_addr)
        self.zs.connect(inspector_zmq_addr) # the inspector should bind it

    def _determine_ports(self, datapath, data):
        ofp = datapath.ofproto
        in_port = ofp.OFPP_CONTROLLER
        out_port = ofp.OFPP_FLOOD
        
        pkt = packet.Packet(data)
        eth = pkt.get_protocols(ethernet.ethernet)[0]
        ### self.mac_to_port is managed by the parent class
        if eth.src in self.mac_to_port[datapath.id]:
            in_port = self.mac_to_port[datapath.id][eth.src]
        if eth.dst in self.mac_to_port[datapath.id]:
            out_port = self.mac_to_port[datapath.id][eth.dst]
        return (in_port, out_port)

    def _zmq_worker(self, datapath):
        while True:
            data = self.zs.recv()
            # from hexdump import hexdump
            # hexdump(data)
            ofp = datapath.ofproto
            parser = datapath.ofproto_parser
            in_port, out_port = self._determine_ports(datapath, data)
            ofp_packet_out = parser.OFPPacketOut(datapath=datapath,
                                                 buffer_id=ofp.OFP_NO_BUFFER,
                                                 in_port=in_port,
                                                 actions=[parser.OFPActionOutput(out_port)],
                                                 data=data)
            # self.logger.debug('Inject: PKT-OUT: %d->%d' % (in_port, out_port))
            datapath.send_msg(ofp_packet_out)
        
    @set_ev_cls(ofp_event.EventOFPSwitchFeatures, CONFIG_DISPATCHER)
    def switch_features_handler(self, ev):
        super(RyuOF13SwitchBase, self).switch_features_handler(ev)
        datapath = ev.msg.datapath
        ofp = datapath.ofproto
        parser = datapath.ofproto_parser

        ### increase miss_send_len to the max (OVS default: 128 bytes)
        conf = parser.OFPSetConfig(datapath, ofp.OFPC_FRAG_NORMAL, 0xFFFF)
        datapath.send_msg(conf)

        ### install self.matches
        actions = [parser.OFPActionOutput(ofp.OFPP_CONTROLLER,
                                          ofp.OFPCML_NO_BUFFER)]
        inst = [parser.OFPInstructionActions(ofp.OFPIT_APPLY_ACTIONS,
                                             actions)]
        for i, match in enumerate(self.matches):
            mod = parser.OFPFlowMod(datapath=datapath,
                             priority = self.FLOW_PRIORITY_BASE + i,
                             match=match,
                             instructions=inst)
            ### TODO: Check flow overlaps
            datapath.send_msg(mod)

        ### initialize mac table (managed by the parent class)
        self.mac_to_port.setdefault(datapath.id, {})

        ### Start ZMQ Worker
        self.logger.info('Starting ZMQ Worker for datapath %d', datapath.id)
        eventlet.spawn(self._zmq_worker, datapath)

        self.logger.info('Setup done for datapath %d', datapath.id)

    @set_ev_cls(ofp_event.EventOFPPacketIn, MAIN_DISPATCHER)
    def _packet_in_handler(self, ev):
        datapath = ev.msg.datapath
        ofp = datapath.ofproto
        parser = datapath.ofproto_parser
        msg = ev.msg
        
        ### NOTE: too old OVS(<= 2.0?) returns OFPR_ACTION instead of OFPR_NO_MATCH
        ### http://git.openvswitch.org/cgi-bin/gitweb.cgi?p=openvswitch;a=commitdiff;h=cfa955b083c5617212a29a03423e063ff6cb350a
        ### So we need OVS >= 2.1.
        if msg.reason == ofp.OFPR_NO_MATCH or DISABLE_INJECTION:
            self.logger.debug('PKT-IN: NOT inject, msg.reason=%d', msg.reason)
            super(RyuOF13SwitchBase, self)._packet_in_handler(ev)
        else:
            self.logger.debug('PKT-IN: inject, msg.reason=%d', msg.reason)
            self.zs.send(msg.data)


import socket            
class RyuOF13Switch(RyuOF13SwitchBase):
    def __init__(self, tcp_ports, udp_ports, zmq_addr, *args, **kwargs):
        OFPMatch = self.OFP_PARSER.OFPMatch
        matches = []
        for port in tcp_ports:
            m = OFPMatch(eth_type=0x0800, ip_proto=socket.IPPROTO_TCP, tcp_dst=port)
            matches.append(m)
            m = OFPMatch(eth_type=0x0800, ip_proto=socket.IPPROTO_TCP, tcp_src=port)
            matches.append(m)
        for port in udp_ports:
            m = OFPMatch(eth_type=0x0800, ip_proto=socket.IPPROTO_UDP, udp_dst=port)
            matches.append(m)
            m = OFPMatch(eth_type=0x0800, ip_proto=socket.IPPROTO_UDP, udp_src=port)
            matches.append(m)
        super(RyuOF13Switch, self).__init__(matches, zmq_addr, *args, **kwargs)

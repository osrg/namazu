from abc import ABCMeta, abstractmethod
import eventlet
from eventlet.green import SocketServer, zmq
import functools
import hexdump
import json
import scapy.all
import six
import requests
from StringIO import StringIO

from .. import LOG as _LOG
from ..entity.entity import *
from ..entity.event import *
from ..entity.action import *
from .ether_tcp_watcher import TCPWatcher

LOG = _LOG.getChild(__name__)

ENABLE_TCP_WATCHER = True

eventlet.monkey_patch()

@six.add_metaclass(ABCMeta)
class EtherInspectorBase(object):
    pkt_recv_handler_table = {}
    def __init__(self, zmq_addr, oc_addr='http://localhost:10000/api/v1', process_id='_earthquake_ether_inspector'):
        if ENABLE_TCP_WATCHER:
            LOG.info('Using TCPWatcher')
            self.tcp_watcher = TCPWatcher()
        else: self.tcp_watcher = None
        self.deferred_events = {} # key: string(event_uuid), value; {'event': PacketEvent, 'packet': PacketBase}

        LOG.info('ZMQ Addr: %s', zmq_addr)        
        self.zmq_addr = zmq_addr
        LOG.info('OC Addr: %s', oc_addr)        
        self.oc_addr = oc_addr
        LOG.info('System Process ID: %s', process_id)                        
        self.process_id = process_id
        

    def start(self):
        zmq_worker_handle = self.start_zmq_worker()
        oc_worker_handle = self.start_oc_worker()
        zmq_worker_handle.wait()
        raise RuntimeError('should not reach here')
        
    def start_zmq_worker(self):
        self.zmq_ctx = zmq.Context()
        self.zs = self.zmq_ctx.socket(zmq.PAIR)
        self.zs.bind(self.zmq_addr)
        worker_handle = eventlet.spawn(self._zmq_worker)
        return worker_handle

    def start_oc_worker(self):
        worker_handle = eventlet.spawn(self._oc_worker)
        return worker_handle

    def regist_layer_on_tcp(self, klazz, tcp_port):
        scapy.all.bind_layers(scapy.all.TCP, klazz, dport=tcp_port)
        scapy.all.bind_layers(scapy.all.TCP, klazz, sport=tcp_port)

    def inspect(self, raw_eth_frame):
        """
        scapy inspector

        Do NOT call TWICE for the same packet, as the inspector can have side-effects
        """
        pkt = scapy.all.Ether(raw_eth_frame) 
        return pkt
                        
    def _zmq_worker(self):
        """
        ZeroMQ worker for the inspector
        """
        while True:
            raw_eth_frame = self.zs.recv()
            try:
                # LOG.info('Full-Hexdump (%d bytes)', len(raw_eth_frame))
                # for line in hexdump.hexdump(raw_eth_frame, result='generator'):
                #     LOG.info(line)
                    
                if self.tcp_watcher:
                    self.tcp_watcher.on_recv(raw_eth_frame, default_handler=self._on_recv_frame_from_switch)
                else:
                    self._on_recv_frame_from_switch(raw_eth_frame)
            except Exception as e:
                LOG.error('Error in _zmq_worker()', exc_info=True)
                try:
                    LOG.error('Full-Hexdump (%d bytes)', len(raw_eth_frame))
                    for line in hexdump.hexdump(raw_eth_frame, result='generator'):
                        LOG.error(line)
                except:
                    LOG.error('Error while hexdumping', exc_info=True)
                self._send_frame_to_switch(raw_eth_frame)

    def _send_frame_to_switch(self, raw_eth_frame):
        self.zs.send(raw_eth_frame)

    def _on_recv_frame_from_switch(self, raw_eth_frame):
        pkt = self.inspect(raw_eth_frame)
        event = self.map_packet_to_event(pkt)
        assert event is None or isinstance(event, PacketEvent)        
        if not event:
            self._send_frame_to_switch(raw_eth_frame)            
        else:
            self.on_packet_event(event, pkt)

    @abstractmethod
    def map_packet_to_event(self, pkt):        
        """
        return None if this packet is NOT interesting at all.
        """
        pass
        
    def on_packet_event(self, ev, pkt, buffer_if_not_sent=False):
        assert isinstance(ev, PacketEvent)
        ev.process = self.process_id
        sent = self.send_event_to_orchestrator(ev)
        if not sent:
            if buffer_if_not_sent:
                LOG.debug('Buffering an event: %s', ev)
            else:
                LOG.debug('Passing an event observed: %s', ev)
                LOG.debug('Pass %s, %s', ev.uuid, pkt.mysummary())
                self._send_frame_to_switch(str(pkt))
                return
        self.defer_packet_event(ev, pkt)
        
    def defer_packet_event(self, ev, pkt):
        """
        Defer the packet until the orchestrator permits
        """
        assert isinstance(ev, PacketEvent)
        assert ev.deferred
        self.deferred_events[ev.uuid] = {'event': ev, 'packet': pkt}
        LOG.debug('Defer event uuid=%s, packet=%s, deferred(after defer)=%d', 
                  ev.uuid, pkt.mysummary(), len(self.deferred_events))
        
    def pass_deferred_event_uuid(self, ev_uuid):
        try:
            event = self.deferred_events[ev_uuid]['event']
            assert isinstance(event, PacketEvent)
            assert event.deferred
            pkt = self.deferred_events[ev_uuid]['packet']
            LOG.debug('Pass deferred event uuid=%s, packet=%s, len(before pass)=%d', 
                      ev_uuid, pkt.mysummary(), len(self.deferred_events))
            self._send_frame_to_switch(str(pkt))
            del self.deferred_events[ev_uuid]
        except Exception as e:
            LOG.error('cannot pass this event: %s', ev_uuid, exc_info=True)

    def send_event_to_orchestrator(self, ev):
        try:
            jsdict = ev.to_jsondict()
            headers = {'content-type': 'application/json'}
            r = requests.post(self.oc_addr, data=json.dumps(jsdict), headers=headers)
            return True
        except Exception as e:
            LOG.error('cannot send event: %s', ev, exc_info=True)
            return False

    def on_recv_action_from_orchestrator(self, action):
        LOG.debug('Received action: %s', action)
        ev_uuid = action.option['event_uuid']
        self.pass_deferred_event_uuid(ev_uuid)

    def _oc_worker(self):
        error_count = 0
        while True:
            try:
                addr = self.oc_addr + '/' + self.process_id
                LOG.debug('GET %s', addr)
                r = requests.get(addr)
                jsdict = r.json()
                action = ActionBase.dispatch_from_jsondict(jsdict)
                self.on_recv_action_from_orchestrator(action)
                error_count = 0
            except Exception as e:
                LOG.error('cannot HTTP GET', exc_info=True)                
                error_count += 1
                eventlet.sleep(error_count * 1.0)

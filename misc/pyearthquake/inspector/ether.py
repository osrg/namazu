from abc import ABCMeta, abstractmethod
import json
import requests

import eventlet
from eventlet.green import zmq, time
import hexdump
import scapy.all
import six

from .. import LOG as _LOG
from ..signal.signal import ActionBase, DEFAULT_ORCHESTRATOR_URL
from ..signal.event import PacketEvent
from ..signal.action import EventAcceptanceAction, PacketFaultAction, NopAction
from .internal.ether_tcp_watcher import TCPWatcher

LOG = _LOG.getChild(__name__)

ENABLE_TCP_WATCHER = True

eventlet.monkey_patch()  # for requests


@six.add_metaclass(ABCMeta)
class EtherInspectorBase(object):
    pkt_recv_handler_table = {}

    def __init__(
        self, zmq_addr, orchestrator_rest_url=DEFAULT_ORCHESTRATOR_URL,
                 entity_id='_namazu_ether_inspector'):
        if ENABLE_TCP_WATCHER:
            LOG.info('Using TCPWatcher')
            self.tcp_watcher = TCPWatcher()
        else:
            self.tcp_watcher = None
        self.deferred_events = {}
            # key: string(event_uuid), value: {'event': PacketEvent,
            # 'metadata`: dict}

        LOG.info('Hookswitch ZMQ Addr: %s', zmq_addr)
        self.zmq_addr = zmq_addr
        LOG.info('Orchestrator REST URL: %s', orchestrator_rest_url)
        self.orchestrator_rest_url = orchestrator_rest_url
        LOG.info('Inspector System Entity ID: %s', entity_id)
        self.entity_id = entity_id

    def start(self):
        zmq_worker_handle = self.start_zmq_worker()
        rest_worker_handle = eventlet.spawn(self._orchestrator_rest_worker)
        zmq_worker_handle.wait()
        rest_worker_handle.wait()
        raise RuntimeError('should not reach here')

    def start_zmq_worker(self):
        self.zmq_ctx = zmq.Context()
        self.zs = self.zmq_ctx.socket(zmq.PAIR)
        self.zs.bind(self.zmq_addr)
        worker_handle = eventlet.spawn(self._zmq_worker)
        return worker_handle

    def regist_layer_on_tcp(self, klazz, tcp_port):
        scapy.all.bind_layers(scapy.all.TCP, klazz, dport=tcp_port)
        scapy.all.bind_layers(scapy.all.TCP, klazz, sport=tcp_port)

    def inspect(self, eth_bytes):
        """
        scapy inspector

        Do NOT call TWICE for the same packet, as the inspector can have side-effects
        """
        pkt = scapy.all.Ether(eth_bytes)
        return pkt

    def _zmq_worker(self):
        """
        ZeroMQ worker for the inspector
        """
        while True:
            metadata_str, eth_bytes = self.zs.recv_multipart()
            metadata = json.loads(metadata_str)
            try:
                # LOG.info('Full-Hexdump (%d bytes)', len(eth_bytes))
                # for line in hexdump.hexdump(eth_bytes, result='generator'):
                #     LOG.info(line)

                if self.tcp_watcher:
                    self.tcp_watcher.on_recv(metadata, eth_bytes,
                                             default_handler=self._on_recv_from_hookswitch,
                                             retrans_handler=self._on_tcp_retrans)
                else:
                    self._on_recv_from_hookswitch(metadata, eth_bytes)
            except Exception as e:
                LOG.error('Error in _zmq_worker()', exc_info=True)
                try:
                    LOG.error('Full-Hexdump (%d bytes)', len(eth_bytes))
                    for line in hexdump.hexdump(eth_bytes, result='generator'):
                        LOG.error(line)
                except:
                    LOG.error('Error while hexdumping', exc_info=True)
                self._send_to_hookswitch(metadata)

    def _send_to_hookswitch(self, metadata, op='accept'):
        assert isinstance(metadata, dict)
        assert op in ('accept', 'drop')
        resp_metadata = metadata.copy()
        resp_metadata['op'] = op
        resp_metadata_str = json.dumps(resp_metadata)
        self.zs.send_multipart((resp_metadata_str, ''))

    def _on_recv_from_hookswitch(self, metadata, eth_bytes):
        inspected_packet = self.inspect(eth_bytes)
        event = self.map_packet_to_event(inspected_packet)
        assert event is None or isinstance(event, PacketEvent)
        if not event:
            self._send_to_hookswitch(metadata, op='accept')
        else:
            self.on_packet_event(metadata, event)

    def _on_tcp_retrans(self, metadata, eth_bytes):
        self._send_to_hookswitch(metadata, op='drop')

    @abstractmethod
    def map_packet_to_event(self, pkt):
        """
        return None if this packet is NOT interesting at all.
        """
        pass

    def on_packet_event(self, metadata, event, buffer_if_not_sent=False):
        assert isinstance(event, PacketEvent)
        event.entity = self.entity_id
        sent = self.send_event_to_orchestrator(event)
        if not sent:
            if buffer_if_not_sent:
                LOG.debug('Buffering an event: %s', event)
            else:
                LOG.debug(
                    'Accepting an event (could not sent to orchestrator): %s', event)
                self._send_to_hookswitch(metadata)
                return
        if event.deferred:
            self.defer_packet_event(metadata, event)
        else:
            # NOTE: non-deferred packet events are useful for logging
            LOG.debug('Accepting an event (not deferred): %s', event)
            self._send_to_hookswitch(metadata)

    def defer_packet_event(self, metadata, event):
        """
        Defer the packet until the orchestrator permits
        """
        assert isinstance(event, PacketEvent)
        assert event.deferred
        LOG.debug('Defer event=%s, deferred+:%d->%d',
                  event, len(self.deferred_events), len(self.deferred_events) + 1)
        self.deferred_events[event.uuid] = {
            'event': event, 'metadata': metadata, 'time': time.time()}

    def complete_deferred_event_uuid(self, event_uuid, op='accept'):
        try:
            event = self.deferred_events[event_uuid]['event']
            assert isinstance(event, PacketEvent)
            assert event.deferred
            metadata = self.deferred_events[event_uuid]['metadata']
            LOG.debug('Complete(%s) deferred event=%s, deferred-:%d->%d',
                      op,
                      event, len(self.deferred_events), len(self.deferred_events) - 1)
            self._send_to_hookswitch(metadata, op)
            del self.deferred_events[event_uuid]
        except Exception as e:
            LOG.error('cannot pass this event: %s', event_uuid, exc_info=True)

    def send_event_to_orchestrator(self, event):
        try:
            event_jsdict = event.to_jsondict()
            headers = {'content-type': 'application/json'}
            post_url = self.orchestrator_rest_url + \
                '/events/' + self.entity_id + '/' + event.uuid
            LOG.debug('POST %s', post_url)
            r = requests.post(
                post_url, data=json.dumps(event_jsdict), headers=headers)
            return True
        except Exception as e:
            LOG.error('cannot send event: %s', event, exc_info=True)
            # do not re-raise the exception to continue processing
            return False

    def on_recv_action_from_orchestrator(self, action):
        LOG.debug('Received action: %s', action)
        if isinstance(action, EventAcceptanceAction):
            self.complete_deferred_event_uuid(action.event_uuid, op='accept')
        elif isinstance(action, PacketFaultAction):
            self.complete_deferred_event_uuid(action.event_uuid, op='drop')
        elif isinstance(action, NopAction):
            LOG.debug('nop action: %s', action)
        else:
            LOG.warn('Unsupported action: %s', action)

    def _orchestrator_rest_worker(self):
        error_count = 0
        got = None
        while True:
            try:
                get_url = self.orchestrator_rest_url + \
                    '/actions/' + self.entity_id
                LOG.debug('GET %s', get_url)
                got = requests.get(get_url)
                got_jsdict = got.json()
                action = ActionBase.dispatch_from_jsondict(got_jsdict)
                LOG.debug('got %s', action.uuid)
                delete_url = get_url + '/' + action.uuid
                LOG.debug('DELETE %s', delete_url)
                deleted = requests.delete(delete_url)
                assert deleted.status_code == 200
                self.on_recv_action_from_orchestrator(action)
                error_count = 0
            except Exception as e:
                LOG.error('cannot HTTP GET', exc_info=True)
                if got is not None:
                    LOG.error('Got: %s', got.text)
                error_count += 1
                eventlet.sleep(error_count * 1.0)
            got = None

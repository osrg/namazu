from abc import ABCMeta, abstractmethod
import eventlet
from eventlet.green import zmq
import six
import json

from pyearthquake import LOG as _LOG

LOG = _LOG.getChild(__name__)

@six.add_metaclass(ABCMeta)
class ZMQClientBase(object):
    def __init__(self, zmq_addr):
        self.zmq_addr = zmq_addr
        self.zmq_ctx = zmq.Context()
        self.zs = self.zmq_ctx.socket(zmq.PAIR)
        self.pendings={} # key: id(int), value: Ethernet Frame

    def start(self):
        LOG.info('Connecting to ZMQ: %s', self.zmq_addr)
        self.zs.connect(self.zmq_addr)
        worker_handle = eventlet.spawn(self._zmq_worker)
        return worker_handle

    def _zmq_worker(self):
        while True:
            # eth_ignored will be used when 'op' == 'modify' is implemented
            metadata_str, eth_ignored = self.zs.recv_multipart()
            metadata = json.loads(metadata_str)
            op = metadata['op']
            assert op in ('accept', 'drop'), 'Invalid op metadata: %s' % (metadata)
            packet_id = metadata['id']
            assert packet_id in self.pendings, 'Unknown packet metadata: %s' %(metadata)
            eth_bytes = self.pendings[packet_id]
            del self.pendings[packet_id]
            LOG.debug('Pendings-(id=%d,op=%s): %d->%d', packet_id, op, len(self.pendings) + 1, len(self.pendings))
            if op == 'accept':
                self.on_accept(packet_id, eth_bytes, metadata)
            elif op == 'drop':
                self.on_drop(packet_id, eth_bytes, metadata)
            else:
                raise RuntimeError('This should not happen')
    
    def send(self, packet_id, eth_bytes):
        metadata = {'id': packet_id}
        metadata_str = json.dumps(metadata)
        self.zs.send_multipart((metadata_str, eth_bytes))
        assert not packet_id in self.pendings
        self.pendings[packet_id] = eth_bytes
        LOG.debug('Pendings+(id=%d): %d->%d', packet_id, len(self.pendings) - 1, len(self.pendings))

    @abstractmethod
    def on_accept(self, packet_id, eth_bytes, metadata=None):
        pass

    @abstractmethod
    def on_drop(self, packet_id, eth_bytes, metadata=None):
        pass

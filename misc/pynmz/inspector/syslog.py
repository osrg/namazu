import SocketServer
from abc import ABCMeta, abstractmethod
import json
import requests

import six

from .. import LOG as _LOG
from ..signal.signal import DEFAULT_ORCHESTRATOR_URL
from ..signal.event import LogEvent

LOG = _LOG.getChild(__name__)


@six.add_metaclass(ABCMeta)
class SyslogInspectorBase(object):

    def __init__(
        self, udp_port=10514, orchestrator_rest_url=DEFAULT_ORCHESTRATOR_URL,
                 entity_id='_namazu_syslog_inspector'):
        LOG.info('Syslog UDP port: %d', udp_port)
        LOG.info('Orchestrator REST URL: %s', orchestrator_rest_url)
        self.orchestrator_rest_url = orchestrator_rest_url
        LOG.info('Inspector System Entity ID: %s', entity_id)
        self.entity_id = entity_id

        that = self

        class SyslogUDPHandler(SocketServer.BaseRequestHandler):

            def handle(self):
                data = bytes.decode(self.request[0].strip(), 'utf-8')
                that.on_syslog_recv(
                    self.client_address[0], self.client_address[1], data)

        self.syslog_server = SocketServer.UDPServer(
            ('0.0.0.0', udp_port), SyslogUDPHandler)

    def start(self):
        self.syslog_server.serve_forever()

    def on_syslog_recv(self, ip, port, data):
        LOG.info('SYSLOG from %s:%d: "%s"', ip, port, data)
        event = self.map_syslog_to_event(ip, port, data)
        assert event is None or isinstance(event, LogEvent)
        if event:
            try:
                self.send_event_to_orchestrator(event)
            except Exception as e:
                LOG.error('cannot send event: %s', event, exc_info=True)

    def send_event_to_orchestrator(self, event):
        event_jsdict = event.to_jsondict()
        headers = {'content-type': 'application/json'}
        post_url = self.orchestrator_rest_url + \
            '/events/' + self.entity_id + '/' + event.uuid
        # LOG.debug('POST %s', post_url)
        r = requests.post(
            post_url, data=json.dumps(event_jsdict), headers=headers)

    @abstractmethod
    def map_syslog_to_event(self, ip, port, data):
        """

        :param ip:
        :param port:
        :param data:
        :return: None or LogEvent
        """
        pass


class BasicSyslogInspector(SyslogInspectorBase):
    # @Override

    def map_syslog_to_event(self, ip, port, data):
        entity = 'entity-%s:%d' % (ip, port)
        event = LogEvent.from_message(entity, data)
        return event


if __name__ == "__main__":
    insp = BasicSyslogInspector()
    insp.start()

from abc import ABCMeta, abstractmethod
import colorama
import copy
import ctypes
import os
import eventlet
from eventlet import wsgi
from eventlet.queue import Queue
from flask import Flask, request, Response, jsonify
import json
import six
import subprocess
import sys
import time
import uuid

from .. import LOG as _LOG
from ..signal.signal import EventBase, ActionBase
from .detector import TerminationDetectorBase
from .digestible import BasicDigestible
from .explorer import ExplorerBase
from .state import BasicState
from .watcher import WatcherBase, DefaultWatcher

LOG = _LOG.getChild('orchestrator.orchestrator')

## these imports are needed for eval(plugin_str)
import pyearthquake
import pyearthquake.orchestrator.orchestrator
import pyearthquake.orchestrator.watcher
import pyearthquake.orchestrator.detector
import pyearthquake.orchestrator.explorer


@six.add_metaclass(ABCMeta)
class OrchestratorBase(object):
    def __init__(self):
        """
        orchestrator plugin class constructor
        """
        LOG.warn('***** pyearthquake orchestrator: UNDER REFACTORING, YOU SHOULD NOT USE THIS (please try v0.1 instead) *****')
        LOG.warn('pyearthquake orchestrator will be provided as a external plugin for earthquake (planned in August 2015)')
        pass

    def _init_early(self):
        self.processes = {}
        self.watchers = []
        self.default_watcher = DefaultWatcher()
        self.default_watcher.init_with_orchestrator(orchestrator=self)

    def init_with_config(self, config):
        """
        orchestrator_loader calls this
        """
        self._init_early()
        self.config = config
        self._init_parse_config()
        self._init_load_libearthquake()
        self._init_load_termination_detector_plugin()
        self._init_load_explorer_plugin()
        self._init_regist_known_processes()

    def regist_process(self, pid):

        assert not pid in self.processes, 'Process %s has been already registered' % pid
        self.processes[pid] = {'actions': [], 'http_action_ready': Queue()}

        def regist_watcher():
            LOG.info('Loading ProcessWatcher "%s"', self.process_watcher_str)
            watcher = eval(self.process_watcher_str)
            assert isinstance(watcher, WatcherBase)
            watcher.init_with_process(orchestrator=self, process_id=pid)
            LOG.info('Loaded ProcessWatcher %s for process %s', watcher, pid)
            self.watchers.append(watcher)  # TODO: integrate self.watchers to self.processes

        regist_watcher()
        LOG.info('Registered Process %s: %s', pid, self.processes[pid])

    def _init_parse_config(self):
        # TODO: define config schema
        self.listen_port = int(self.config['globalFlags']['orchestratorListenPort'])
        self.process_watcher_str = self.config['globalFlags']['plugin']['processWatcher']
        self.explorer_str = self.config['globalFlags']['plugin']['explorer']
        self.detector_str = self.config['globalFlags']['plugin']['terminationDetector']

        if 'processes' in self.config:
            self.known_processes = self.config['processes']
        else:
            self.known_processes = []

    def _init_load_libearthquake(self):
        dll_str = 'libearthquake.so'
        LOG.info('Loading DLL "%s"', dll_str)
        self.libearthquake = ctypes.CDLL(dll_str)
        LOG.info('Loaded DLL "%s"', self.libearthquake)
        config_json_str = json.dumps(self.config)
        rc = self.libearthquake.EQInitCtx(config_json_str)
        assert rc == 0

    def _init_load_explorer_plugin(self):
        LOG.info('Loading explorer "%s"', self.explorer_str)
        self.explorer = eval(self.explorer_str)
        assert isinstance(self.explorer, ExplorerBase)
        LOG.info('Loaded explorer %s', self.explorer)
        initial_state = self.make_initial_state()
        self.explorer.init_with_orchestrator(self, initial_state)

    def _init_load_termination_detector_plugin(self):
        LOG.info('Loading termination detector "%s"', self.detector_str)
        self.termination_detector = eval(self.detector_str)
        assert isinstance(self.termination_detector, TerminationDetectorBase)
        LOG.info('Loaded termination detector %s', self.termination_detector)
        self.termination_detector.init_with_orchestrator(self)

    def _init_regist_known_processes(self):
        for p in self.known_processes:
            LOG.debug('Registering a known process: %s', p)
            pid = p['id']
            self.regist_process(pid)

    def start(self):
        explorer_worker_handle = eventlet.spawn(self.explorer.worker)
        flask_app = Flask(self.__class__.__name__)
        flask_app.debug = True
        self.regist_flask_routes(flask_app)
        sock = eventlet.listen(('localhost', self.listen_port))
        wsgi.server(sock, flask_app)
        explorer_worker_handle.wait()
        raise RuntimeError('should not reach here! (TODO: support shutdown on inspection termination)')

    def regist_flask_routes(self, app):
        LOG.debug('registering flask routes')
        api_root = '/api/v2'
        LOG.debug('REST API root=%s', api_root)

        @app.route('/', methods=['GET'])
        def GET_root():
            text = 'Hello Earthquake! -- %s(pid=%d)\n' % (str(self), os.getpid())
            return Response(text, mimetype='text/plain')

        @app.route('/api/v1', methods=['GET', 'POST'])
        def GETPOST_old_api_v1():
            """
            API v1 has been eliminated. So return 406 Not Acceptable
            """
            return Response('API v1 has been eliminated. Use API v2 or later.\n', status=406, mimetype='text/plain')

        @app.route(api_root + '/ctrl/force_terminate', methods=['POST'])
        def POST_ctrl_force_terminate():
            """
            Forcibly terminate inspection
            """
            # We don't support GET, as GET must be idempotent (RFC 7231)
            state = self.explorer.state
            LOG.debug('Calling state.force_terminate()')
            state.force_terminate()
            return jsonify({})

        @app.route(api_root + '/visualizers/csv', methods=['GET'])
        def GET_visualizers_csv():
            """
            Visualize (to be eliminated?)
            """
            csv_fn = self.libearthquake.EQGetStatCSV_UnstableAPI
            csv_fn.restype = ctypes.c_char_p
            csv = csv_fn()
            LOG.debug('CSV <== %s', csv)
            return Response(csv, mimetype='text/csv')


        @app.route(api_root + '/events/<entity_id>/<event_uuid>', methods=['POST'])
        def POST_events(process_id, event_uuid):
            ## get event
            ## NOTE: get_json(force=True): ignore mimetype('application/json')
            event_jsdict = request.get_json(force=True)
            LOG.debug('API ==> %s', event_jsdict)
            assert event_jsdict['uuid'] == event_uuid

            ## send event to explorer
            event = EventBase.dispatch_from_jsondict(event_jsdict)
            event.recv_timestamp = time.time()
            self.explorer.send_event(event)
            return jsonify({})

        @app.route(api_root + '/actions/<entity_id>', methods=['GET'])
        def GET_actions(process_id):
            ## regist process, if not registed
            if not process_id in self.processes:
                self.regist_process(process_id)

            def wait_for_actions():
                while True:
                    actions_len = len(self.processes[process_id]['actions'])
                    LOG.debug('#actions=%d', actions_len)
                    if actions_len > 0: break
                    LOG.debug('waiting for a new action')
                    self.processes[process_id]['http_action_ready'].get()
                return self.processes[process_id]['actions'][0]

            action = wait_for_actions()

            ## return action
            action_jsdict = action.to_jsondict()
            LOG.debug('API <== %s', action_jsdict)
            return jsonify(action_jsdict)

        @app.route(api_root + '/actions/<entity_id>/<action_uuid>', methods=['DELETE'])
        def DELETE_actions(process_id, action_uuid):
            assert process_id in self.processes
            actions = [(i, x) for i, x in enumerate(self.processes[process_id]['actions']) if x.uuid == action_uuid]
            assert len(actions) <= 1
            if len(actions) == 0:
                ## this is expected because DELETE must be idempotent
                LOG.warn('Action %s has been already deleted?', action_uuid)
                return jsonify({})
            i = actions[0][0]
            del self.processes[process_id]['actions'][i]
            LOG.debug('DELETE %s', action_uuid)
            return jsonify({})

    def send_action(self, action):
        """
        explorer calls this
        """
        process_id = action.process
        self.processes[process_id]['actions'].append(action)
        self.processes[process_id]['http_action_ready'].put(True)
        LOG.debug('Enqueued action %s, #actions=%d', action, len(self.processes[process_id]['actions']))

    def execute_command(self, command):
        rc = subprocess.call(command, shell=True)
        return rc

    @abstractmethod
    def call_action(self, action):
        """
        it may be interesting to override this
        """
        pass

    @abstractmethod
    def make_digestible_pair(self, event, action):
        """
        it may be interesting to override this
        """
        pass

    @abstractmethod
    def make_initial_state(self):
        """
        it may be interesting to override this
        """
        pass


class BasicOrchestrator(OrchestratorBase):
    def call_action(self, action):
        assert isinstance(action, ActionBase)
        action.call(orchestrator=self)

    def make_digestible_pair(self, event, action):
        digestible = BasicDigestible(event, action)
        return digestible

    def make_initial_state(self):
        return BasicState()

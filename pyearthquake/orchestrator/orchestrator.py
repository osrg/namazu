from abc import ABCMeta, abstractmethod
import colorama
import copy
import ctypes
import eventlet
from eventlet import wsgi
from eventlet.queue import *
from flask import Flask, request, Response, jsonify
import json
import six
import subprocess
import sys
import time
import uuid


from .. import LOG as _LOG
from ..entity import *
from ..entity.entity import *
from ..util import *


from .digestible import *
from .state import *
from .watcher import *
from .detector import *
from .explorer import *

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
        pass
    
    def init_with_config(self, config):
        self.config = config
        self._init_parse_config()
        self._init_load_libearthquake()
        self._init_load_process_watcher_plugin()
        self._init_load_termination_detector_plugin()                
        self._init_load_explorer_plugin()
        
    def _init_parse_config(self):
        self.listen_port = int(self.config['globalFlags']['orchestratorListenPort'])
        self.processes = {}
        for p in self.config['processes']:
            pid = p['id']
            queue = Queue()
            self.processes[pid] = { 'queue': queue }
        LOG.info('Processes: %s', self.processes)
        
    def _init_load_libearthquake(self):
        dll_str = 'libearthquake.so'
        LOG.info('Loading DLL "%s"', dll_str)        
        self.libearthquake = ctypes.CDLL(dll_str)
        LOG.info('Loaded DLL "%s"', self.libearthquake)        
        config_json_str = json.dumps(self.config)
        rc = self.libearthquake.EQInitCtx(config_json_str)
        assert rc == 0

    def _init_load_process_watcher_plugin(self):
        self.watchers = []
        self.default_watcher = DefaultWatcher()
        self.default_watcher.init_with_orchestrator(orchestrator=self)
        watcher_str = self.config['globalFlags']['plugin']['processWatcher']
        LOG.info('Loading ProcessWatcher "%s"', watcher_str)
        for p in self.processes:
            watcher = eval(watcher_str)
            assert isinstance(watcher, WatcherBase)
            watcher.init_with_process(orchestrator=self, process_id=p)
            LOG.info('Loaded ProcessWatcher %s for process %s', watcher, p)                        
            self.watchers.append(watcher)
        
    def _init_load_explorer_plugin(self):
        explorer_str = self.config['globalFlags']['plugin']['explorer']
        LOG.info('Loading explorer "%s"', explorer_str)
        self.explorer = eval(explorer_str)
        assert isinstance(self.explorer, ExplorerBase)        
        LOG.info('Loaded explorer %s', self.explorer)            
        self.explorer.init_with_orchestrator(self, initial_state=BasicState())

    def _init_load_termination_detector_plugin(self):
        detector_str = self.config['globalFlags']['plugin']['terminationDetector']
        LOG.info('Loading termination detector "%s"', detector_str)
        self.termination_detector = eval(detector_str)
        assert isinstance(self.termination_detector, TerminationDetectorBase)
        LOG.info('Loaded termination detector %s', self.termination_detector)            
        self.termination_detector.init_with_orchestrator(self)
        
    def start(self):
        explorer_worker_handle = eventlet.spawn(self.explorer.worker)
        flask_app = Flask(self.__class__.__name__)
        flask_app.debug = True
        self.regist_flask_routes(flask_app)
        server_sock = eventlet.listen(('localhost', self.listen_port))
        wsgi.server(server_sock, flask_app)
        raise RuntimeError('should not reach here!')

    def regist_flask_routes(self, app):
        LOG.debug('registering flask routes')
        @app.route('/')
        def root():
            return 'Hello Earthquake!'

        @app.route('/visualize_api/csv', methods=['GET'])
        def visualize_api_csv():
            csv_fn = self.libearthquake.EQGetStatCSV_UnstableAPI
            csv_fn.restype = ctypes.c_char_p
            csv_str = csv_fn()
            LOG.debug('CSV <== %s', csv_str)
            return Response(csv_str, mimetype='text/csv')
        
        @app.route('/api/v1', methods=['POST'])
        def api_v1_post():
            ## get event
            ev_jsdict = request.get_json(force=True)
            LOG.debug('API ==> %s', ev_jsdict)            

            ## check process id (TODO: check dup)
            process_id = ev_jsdict['process']
            assert process_id in self.processes.keys(), 'unknown process %s. check the config.' % (process_id)

            ## send event to explorer
            ev = EventBase.dispatch_from_jsondict(ev_jsdict)
            ev.recv_timestamp = time.time()
            self.explorer.send_event(ev)
            return jsonify({})


        @app.route('/api/v1/<process_id>', methods=['GET'])
        def api_v1_get(process_id):
            assert process_id in self.processes.keys(), 'unknown process %s. check the config.' % (process_id)
             
            ## wait for action from explorer
            got = self.processes[process_id]['queue'].get()
            action = got['action']
            assert isinstance(action, ActionBase)

            ## return action
            action_jsdict = action.to_jsondict()
            LOG.debug('API <== %s', action_jsdict)
            return jsonify(action_jsdict)
    
    def send_action(self, action):
        """
        explorer calls this
        """
        process_id = action.process
        self.processes[process_id]['queue'].put({'type': 'action', 'action': action})
        
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


class BasicOrchestrator(OrchestratorBase):
    def call_action(self, action):
        assert isinstance(action, ActionBase)
        action.call(orchestrator=self)

    def make_digestible_pair(self, event, action):
        digestible = BasicDigestible(event, action)
        return digestible

from abc import ABCMeta, abstractmethod
import colorama
import copy
import ctypes
import eventlet
from eventlet import wsgi
from eventlet.semaphore import *
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

    def _init_early(self):
        self.processes = {}        
        self.watchers = []
        self.default_watcher = DefaultWatcher()
        self.default_watcher.init_with_orchestrator(orchestrator=self)
    
    def init_with_config(self, config):
        self._init_early()
        self.config = config
        self._init_parse_config()
        self._init_load_libearthquake()
        self._init_load_termination_detector_plugin()                
        self._init_load_explorer_plugin()
        self._init_regist_known_processes()

    def regist_process(self, pid):
        queue = Queue()
        sem = Semaphore()
        self.processes[pid] = { 'queue': queue, 'sem': sem }

        def regist_watcher():
            LOG.info('Loading ProcessWatcher "%s"', self.process_watcher_str)
            watcher = eval(self.process_watcher_str)
            assert isinstance(watcher, WatcherBase)
            watcher.init_with_process(orchestrator=self, process_id=pid)
            LOG.info('Loaded ProcessWatcher %s for process %s', watcher, pid)                        
            self.watchers.append(watcher) # TODO: integrate self.watchers to self.processes

        regist_watcher()
        LOG.info('Registered Process %s: %s', pid, self.processes[pid])
        
    def _init_parse_config(self):
        self.listen_port = int(self.config['globalFlags']['orchestratorListenPort'])
        self.process_watcher_str = self.config['globalFlags']['plugin']['processWatcher']
        self.explorer_str = self.config['globalFlags']['plugin']['explorer']
        self.detector_str = self.config['globalFlags']['plugin']['terminationDetector']
        
        self.known_processes = self.config['processes']

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
            assert self.validate_process_id(process_id)                 
            if not process_id in self.processes:
                self.regist_process(process_id)

            ## send event to explorer
            ev = EventBase.dispatch_from_jsondict(ev_jsdict)
            ev.recv_timestamp = time.time()
            self.explorer.send_event(ev)
            return jsonify({})


        @app.route('/api/v1/<process_id>', methods=['GET'])
        def api_v1_get(process_id):
            ## FIXME: make sure single conn exists for a single <process_id>
            assert self.validate_process_id(process_id)
            if not process_id in self.processes:
                self.regist_process(process_id)

            ## acquire sem
            sem_acquired = self.processes[process_id]['sem'].acquire(blocking=False)
            if not sem_acquired:
                err = 'Could not acquire semaphore for %s' % process_id
                LOG.warn(err)
                return err
             
            ## wait for action from explorer
            ## (FIXME: we should break this wait and release the sem when the conn is closed)
            got = self.processes[process_id]['queue'].get()
            action = got['action']
            LOG.debug('Dequeued action %s', action)            
            assert isinstance(action, ActionBase)

            ## return action
            action_jsdict = action.to_jsondict()
            LOG.debug('API <== %s', action_jsdict)

            ## release sem
            self.processes[process_id]['sem'].release()
            return jsonify(action_jsdict)
    
    def send_action(self, action):
        """
        explorer calls this
        """
        process_id = action.process
        # no need to acquire sem
        self.processes[process_id]['queue'].put({'type': 'action', 'action': action})
        LOG.debug('Enqueued action %s', action)
        
    def execute_command(self, command):
        rc = subprocess.call(command, shell=True)
        return rc

    def validate_process_id(self, process_id):
        # TODO: check dup
        return True

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

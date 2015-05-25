from abc import ABCMeta, abstractmethod
import colorama
import copy
import six
import time
import uuid

from .. import LOG as _LOG
from ..entity import *
from ..entity.event import *
from ..entity.action import *
from ..util import *
from .digestible import *

LOG = _LOG.getChild('orchestrator.watcher')

@six.add_metaclass(ABCMeta)
class WatcherBase(object):
    def __init__(self):
        pass

    def init_with_orchestrator(self, orchestrator):
        self.oc = orchestrator
        
    @abstractmethod
    def handles(self, event):
        pass

    @abstractmethod
    def on_event(self, state, event):
        """
        returns list of Digestible pair
        """
        pass

    def on_terminal_state(self, terminal_state):
        pass

    def on_reset(self):
        pass

        
class DefaultWatcher(WatcherBase):
    
    def handles(self, event):
        raise RuntimeError('Do not call handles() for DefaultWatcher.' + 
                           'DefaultWatcher handles the event if no watcher handles it.')

    def on_event(self, state, event):
        if event.deferred:
            LOG.warn('DefaultWatcher passing %s immediately. the event is not passed to the explorer', event)
            action = PassDeferredEventAction.from_event(event)
            self.oc.call_action(action)
        else:
            LOG.warn('DefaultWatcher ignoring %s', event)
        return []



class BasicProcessWatcher(WatcherBase):
    def init_with_process(self, orchestrator, process_id):
        self.init_with_orchestrator(orchestrator)
        self.process = process_id

    def handles(self, event):
        return event.process == self.process

    def on_event(self, state, event):
        pairs = []
        if event.deferred:
            action = PassDeferredEventAction.from_event(event)
        else:
            action = NopAction.from_event(event)
        pair = self.oc.make_digestible_pair(event, action)
        pairs.append(pair)
        # you can override this to add ExecuteCommandActions.        
        # you can add safety checks (i.e., assertion) here
        return pairs

    def on_terminal_state(self, terminal_state):
        LOG.info('Process %s: terminated', self.process)
        
    def on_reset(self):
        LOG.info('Process %s: please restart me manually', self.process)

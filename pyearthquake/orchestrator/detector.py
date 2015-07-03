from abc import ABCMeta, abstractmethod
import copy
import six
import time

from .. import LOG as _LOG
from ..entity import *
from ..util import *
from .digestible import *

LOG = _LOG.getChild('orchestrator.detector')
    
@six.add_metaclass(ABCMeta)
class TerminationDetectorBase(object):
    def init_with_orchestrator(self, orchestrator):
        self.oc = orchestrator
        
    @abstractmethod
    def is_terminal_state(self, state):
        pass


class IdleForWhileDetector(TerminationDetectorBase):
    def __init__(self, msecs=5000):
        self.threshold_msecs = msecs
        
    def is_terminal_state(self, state):
        now = time.time()
        elapsed_secs = now - state.last_transition_time
        elapsed_msecs = elapsed_secs * 1000
        if elapsed_msecs > self.threshold_msecs and  len(state.digestible_sequence) > 0:
            LOG.debug('%s detected termination, as elapsed_msecs=%f > %d', self.__class__.__name__, elapsed_msecs, self.threshold_msecs)
            return True
        return False


class InspectionEndDetector(TerminationDetectorBase):
    """
    detect termination when InspectionEndEvent from all processes are observed
    """
    def is_terminal_state(self, state):
        """
        FIXME: make me light-weight
        """
        
        process_ended = {}
        for pid in self.oc.processes.keys():
            process_ended[pid] = False

        for d in state.digestible_sequence:
            if isinstance(d.event, InspectionEndEvent):
                pid = d.event.process
                process_ended[pid] = True
        
        terminated = process_ended.values() == [True] * len(process_ended)

        if terminated:
            LOG.debug("%s detected terminated=%s", self.__class__.__name__, terminated)        
        return terminated


class ForciblyInspectionEndDetector(TerminationDetectorBase):
    def is_terminal_state(self, state):
        if state.forcibly_inspection_ended is None:
            return False
        return state.forcibly_inspection_ended

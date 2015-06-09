from abc import ABCMeta, abstractmethod
import copy
import six
import time

from .. import LOG as _LOG
from ..entity import *
from ..util import *
from .digestible import *

LOG = _LOG.getChild('orchestrator.state')
    
@six.add_metaclass(ABCMeta)
class StateBase(object):
    def __init__(self):
        self.digestible_sequence = []
        self.init_time = time.time()
        self.last_transition_time = 0
        
    def __repr__(self):
        return '<State %s>' % repr(self.digestible_sequence)

    def __str__(self):
        """
        human-readable representation
        """
        return self.__repr__()
        
    def __hash__(self):
        """
        FIXME: https://github.com/osrg/earthquake/issues/5
        """
        filtered = filter(lambda d: not isinstance(d.event, InspectionEndEvent), self.digestible_sequence)
        return hash(tuple(filtered))

    def __eq__(self, other):
        """
        needed for networkx
        """
        return hash(self) == hash(other)

    def __ne__(self, other):
        return not self.__eq__(other)

    def to_short_str(self):
        return '<State hash=0x%x>' % hash(self)

    def to_jsondict(self):
        ## NOTE: bare list such as '[{"foo":"foo_value"}, {"bar":"bar_value"} ]' is NOT a legal json string. 
        ## JQ cannot handle such an illegal json string.
        return {
            'type': 'list', 
            'elements': [d.to_jsondict() for d in self.digestible_sequence]
        }

    def make_copy(self):
        try:
            copied = copy.copy(self)
            copied.digestible_sequence = copy.copy(self.digestible_sequence)
            return copied
        except Exception as e:
            LOG.error('make_copy() failed for %s', self)
            raise e

    def append_digestible(self, d):
        assert isinstance(d, DigestibleBase)
        self.digestible_sequence.append(d)
        self.last_transition_time = time.time()


class BasicState(StateBase):
    pass


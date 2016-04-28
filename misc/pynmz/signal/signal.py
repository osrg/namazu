from abc import ABCMeta
import six
import uuid
from .. import LOG as _LOG
LOG = _LOG.getChild('signal.signal')


API_ROOT = '/api/v3'
DEFAULT_ORCHESTRATOR_URL = 'http://localhost:10080' + API_ROOT

@six.add_metaclass(ABCMeta)
class SignalBase(object):
    # included in json ('type')
    type_name = '_meta'

    # these are included in json
    entity = '_namazu_invalid_entity_id'
    uuid = '00000000-0000-0000-0000-000000000000'
    option = {}

    # extra variables can be included in json (e.g. 'deferred')
    var_names = ['entity', 'uuid', 'option']

    # used for dispatching
    _children = {}

    def __init__(self):
        self.uuid = str(uuid.uuid4())  # random uuid

    def __str__(self):
        return self.__repr__()

    def __repr__(self):
        return str(self.to_jsondict())

    def to_jsondict(self):
        jsdict = {
            'type': self.type_name,
            'class': self.__class__.__name__,
        }
        for var_name in self.var_names:
            var_val = getattr(self, var_name)
            jsdict[var_name] = var_val
        return jsdict

    def parse_jsondict(self, jsdict):
        assert jsdict['type'] == self.type_name
        for var_name in self.var_names:
            try:
                var_val = jsdict[var_name]
            except KeyError:
                var_val = None
            setattr(self, var_name, var_val)

    @classmethod
    def from_jsondict(cls, jsdict):
        if jsdict['class'] != cls.__name__:
            raise ValueError('%s != %s' % (jsdict['class'], cls.__name__))
        try:
            inst = cls()
        except TypeError, e:
            LOG.error(
                '%s() should not take any mandatory arg other than self', cls)
            raise e
        inst.parse_jsondict(jsdict)
        return inst

    @classmethod
    def register_for_dispatch(cls, klazz):
        cls._children[klazz.__name__] = klazz

    @classmethod
    def dispatch_from_jsondict(cls, jsdict):
        try:
            class_name = jsdict['class']
        except KeyError as e:
            raise cls.RegistryError(
                'Registry not found for type_name %s. (%s)' % (cls.type_name, jsdict))
        try:
            klazz = cls._children[class_name]
        except KeyError as e:
            raise cls.RegistryError(
                'Registry not found for class %s. (%s)' % (class_name, jsdict))
        return klazz.from_jsondict(jsdict)

    class RegistryError(Exception):
        pass

    @classmethod
    def deco(cls):
        def f(klazz):
            cls.register_for_dispatch(klazz)
            return klazz
        return f


@six.add_metaclass(ABCMeta)
class EventBase(SignalBase):
    """
    Event:  Inspector --> Orchestrator
    """
    type_name = 'event'
    deferred = False
    replay_hint = ''
    var_names = SignalBase.var_names + ['deferred', 'replay_hint']
    # orchestrator sets recv_timestamp
    recv_timestamp = -1

event_class = EventBase.deco


@six.add_metaclass(ABCMeta)
class ActionBase(SignalBase):
    """
    Action:  Inspector <-- Orchestrator
    """
    type_name = 'action'
    # event_uuid is not mandatory
    event_uuid = None
    var_names = SignalBase.var_names + ['event_uuid']

action_class = ActionBase.deco

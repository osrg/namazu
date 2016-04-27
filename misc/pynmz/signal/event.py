from .signal import EventBase, event_class
from .. import LOG as _LOG
LOG = _LOG.getChild('signal.event')


@event_class()
class PacketEvent(EventBase):

    """
    L7 packet message
    """
    deferred = True

    @classmethod
    def from_message(cls, src_entity, dst_entity, message, replay_hint=""):
        inst = cls()
        # we do not set inst.entity here
        inst.option = {
            'src_entity': src_entity,
            'dst_entity': dst_entity,
            'message': message
        }
        inst.replay_hint = replay_hint
        return inst


@event_class()
class LogEvent(EventBase):

    """
    syslog (not deferrable)
    """
    deferred = False

    @classmethod
    def from_message(cls, src_entity, message):
        inst = cls()
        # we do not set inst.entity here
        inst.option = {
            'src_entity': src_entity,
            'message': message
        }
        return inst

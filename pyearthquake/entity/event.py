from .entity import EventBase, event_class
from .. import LOG as _LOG
LOG = _LOG.getChild('entity.event')

@event_class()
class FunctionCallEvent(EventBase):
    """
    function call
    """
    deferred = True
    def parse_jsondict(self, jsdict):
        assert 'func_name' in jsdict['option'], 'func_name required'
        super(FunctionCallEvent, self).parse_jsondict(jsdict)

        
@event_class()
class PacketEvent(EventBase):
    """
    L7 packet message
    """
    deferred = True

    @classmethod
    def from_message(cls, src_process, dst_process, message):
        inst = cls()
        # we do not set inst.process here
        inst.option = {
            'src_process': src_process,
            'dst_process': dst_process,
            'message': message
        }                
        return inst


@event_class()
class LogEvent(EventBase):
    """
    syslog (not deferrable)
    """
    deferred = False


@event_class()
class InspectionEndEvent(EventBase):
    """
    Inspection end (not deferrable)
    """
    deferred = False

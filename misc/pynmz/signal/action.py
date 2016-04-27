from .signal import ActionBase, action_class
from .. import LOG as _LOG
LOG = _LOG.getChild('signal.action')


@action_class()
class NopAction(ActionBase):
    pass

@action_class()
class EventAcceptanceAction(ActionBase):
    pass

@action_class()
class PacketFaultAction(ActionBase):
    pass

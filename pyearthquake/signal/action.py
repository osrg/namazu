from .signal import ActionBase, action_class
from .. import LOG as _LOG
LOG = _LOG.getChild('signal.action')

@action_class()
class NopAction(ActionBase):
    @classmethod
    def from_event(cls, event):
        inst = cls()
        inst.entity = event.entity
        return inst

@action_class()
class AcceptDeferredEventAction(ActionBase):
    @classmethod
    def from_event(cls, event):
        assert event.deferred
        inst = cls()
        inst.entity = event.entity
        inst.option = {'event_uuid': event.uuid}                
        return inst

    def digest(self):    
        return (self.__class__.__name__)

# TODO: implement DropDeferredEventAction

@action_class()
class ExecuteCommandOnInspectorAction(ActionBase):
    def __init__(self):
        raise NotImplementedError


@action_class()
class ExecuteCommandOnOrchestratorAction(ActionBase):
    """
    Execute the command on orchestrator, not on inspector.

    This action is recommended for fault injection that kills the inspector.

    """
    def __init__(self):
        raise NotImplementedError

    # @classmethod
    # def from_command(cls, command):
    #     inst = cls()
    #     inst.option = {'command': command}
    #     return inst
    #
    # def call(self, orchestrator):
    #     command = self.option['command']
    #     LOG.debug('%s: execute command"%s"', self.__class__.__name__, command)
    #     rc = orchestrator.execute_command(command)
    #     LOG.debug('%s: return command="%s", rc=%d', self.__class__.__name__, command, rc)

from abc import ABCMeta, abstractmethod
import six

from .. import LOG as _LOG
from ..entity.entity import EventBase, ActionBase

LOG = _LOG.getChild('orchestrator.digestible')


@six.add_metaclass(ABCMeta)
class DigestibleBase(object):
    """
    digestible pair of <Event, Action>.
    digest is used for recording and comparison of history 
    (see to_jsondict())
    """

    def __init__(self, event, action):
        """
        :param event: EventBase
        :param action: ActionBase
        :return: None
        """
        assert event.process == action.process
        self.event = event
        self.action = action

    @abstractmethod
    def digest_event(self):
        pass

    @abstractmethod
    def digest_action(self):
        pass

    def to_jsondict(self):
        """
        makes digest for recording history of executed Pair<Event, Action>.
        digest should NOT include any kind of timestamps, random numbers, UUIDs, or so on.
        """
        d = {
            'process': self.event.process,
            'event_digest': self.digest_event(),
            'action_digest': self.digest_action(),
        }
        return d

    def __repr__(self):
        return '<Digestible %s>' % repr(self.to_jsondict())

    def __str__(self):
        return repr(self)

    def __hash__(self):
        """
        FIXME: https://github.com/osrg/earthquake/issues/5
        """
        return hash(repr(self.to_jsondict()))

    def __eq__(self, other):
        return hash(self) == hash(other)

    def __ne__(self, other):
        return not self.__eq__(other)


class BasicDigestible(DigestibleBase):
    def digest_event(self):
        assert isinstance(self.event, EventBase)
        return self.event.digest()

    def digest_action(self):
        assert isinstance(self.action, ActionBase)
        return self.action.digest()

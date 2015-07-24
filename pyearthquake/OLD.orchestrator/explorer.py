from abc import ABCMeta, abstractmethod
import colorama
import random
import json

from eventlet.greenthread import sleep
from eventlet.timeout import Timeout
from eventlet.queue import *
import six
import time

from .. import LOG as _LOG
from ..entity.entity import EventBase, ActionBase
from .digestible import DigestibleBase

LOG = _LOG.getChild('orchestrator.explorer')


@six.add_metaclass(ABCMeta)
class ExplorerBase(object):
    def __init__(self):
        # self.graph = None
        self._event_q = Queue()
        self.oc = None
        self.state = None
        self.initial_state = None
        self.visited_terminal_states = {}  # key: state, value: count (TODO: MOVE TO LIBEARTHQUAKE.SO)
        self.time_slice = 0

    def init_with_orchestrator(self, oc, initial_state):
        """
        :param oc: OrchestratorBase
        :param initial_state: StateBase
        :return: None
        """
        self.oc = oc
        self.initial_state = initial_state
        self.state = self.initial_state.make_copy()
        LOG.debug(colorama.Back.BLUE +
                  'set initial state=%s' +
                  colorama.Style.RESET_ALL, self.state.to_short_str())
        # self.graph = Graph(self.state)

    def send_event(self, event):
        """
        Send event *to* explorer
        :param event: EventBase
        :return: None
        """
        assert isinstance(event, EventBase)
        self._event_q.put(event)

    def recv_events(self, timeout_msecs):
        """
        Let explorer receive events
        :param timeout_msecs: int
        :return:
        """
        events = []
        timeout = Timeout(timeout_msecs / 1000.0)
        try:
            while True:
                event = self._event_q.get()
                events.append(event)
        except Timeout:
            pass
        except Exception as e:
            raise e
        finally:
            timeout.cancel()
        return events

    def _worker__print_events_and_digestibles(self, digestibles, new_events, new_digestibles):
        if digestibles:
            LOG.debug('Before state %s, the following OLD %d digestibles had been yielded', self.state.to_short_str(),
                      len(digestibles))
            for digestible in digestibles: LOG.debug('* %s', digestible)
        LOG.debug('In state %s, the following %d events happend', self.state.to_short_str(), len(new_events))
        for e in new_events:
            try:
                LOG.debug('* %f: %s', e.recv_timestamp, e.abstract_msg)
            except Exception:
                LOG.debug('* %s', e)
        LOG.debug('In state %s, the following NEW %d digestibles were yielded for the above %d events',
                  self.state.to_short_str(), len(new_digestibles), len(new_events))
        for new_digestible in new_digestibles: LOG.debug('* %s', new_digestible)

    def worker(self):
        digestibles = []
        while True:
            if self.oc.termination_detector.is_terminal_state(self.state): self.state = self.on_terminal_state()
            new_events = self.recv_events(timeout_msecs=self.time_slice)
            if not new_events and not digestibles: continue

            new_digestibles = []
            for e in new_events:
                e_handled = False
                for w in self.oc.watchers:
                    if w.handles(e): new_digestibles.extend(w.on_event(self.state, e)); e_handled = True
                if not e_handled: new_digestibles.extend(self.oc.default_watcher.on_event(self.state, e))
            self._worker__print_events_and_digestibles(digestibles, new_events, new_digestibles)
            digestibles.extend(new_digestibles)
            if not digestibles: LOG.warn('No DIGESTIBLE, THIS MIGHT CAUSE FALSE DEADLOCK, state=%s',
                                         self.state.to_short_str())

            next_state, digestibles = self.do_it(digestibles)
            if not digestibles: LOG.warn('No DIGESTIBLE, THIS MIGHT CAUSE FALSE DEADLOCK, next_state=%s',
                                         next_state.to_short_str())

            LOG.debug('transit from %s to %s', self.state.to_short_str(), next_state.to_short_str())
            self.state = next_state

    def do_it(self, digestibles):
        """
        select a digestible from digestibles and do it in the state.
        returns: (next_state, other_digestibles)
        FIXME: rename me!
        """
        if not digestibles: return self.state, []
        chosen_digestible = self.choose_digestible(digestibles)
        LOG.debug('Chosen digestible: %s', chosen_digestible)
        assert (any(digestible.event.uuid == chosen_digestible.event.uuid for digestible in digestibles))
        digestibles_len_before_remove = len(digestibles)
        digestibles.remove(chosen_digestible)
        assert len(digestibles) == digestibles_len_before_remove - 1, 'hash race?'
        other_digestibles = digestibles

        if chosen_digestible:
            next_state = self.do_transition(chosen_digestible)
        else:
            LOG.warn('No DIGESTIBLE chosen, THIS MIGHT CAUSE FALSE DEADLOCK, state=%s', self.state.to_short_str())
            next_state = self.state

        ## NOTE: as other digestibles are also enabled in the NEXT state, we return other digestibles here.
        ## the worker will handle other digestibles in the next round.
        return next_state, other_digestibles

    @abstractmethod
    def choose_digestible(self, digestibles):
        pass

    def call_action(self, action):
        self.oc.call_action(action)

    def do_transition(self, digestible):
        assert isinstance(digestible, DigestibleBase)
        LOG.debug(colorama.Back.BLUE +
                  "Invoking the action:\n" +
                  "  action=%s\n" +
                  "  event=%s\n" +
                  "  state=%s\n" +
                  "  digestible=%s\n" +
                  colorama.Style.RESET_ALL,
                  digestible.action, digestible.event,
                  self.state.to_short_str(),
                  digestible)
        self.call_action(digestible.action)
        next_state = self.state.make_copy()
        next_state.append_digestible(digestible)
        LOG.debug(colorama.Back.BLUE +
                  'State Transition: %s->%s' +
                  colorama.Style.RESET_ALL, self.state.to_short_str(), next_state.to_short_str())

        # self.graph.visit_edge(self.state, next_state, digestible)
        ## NOTE: worker sets self.state to next_state
        return next_state

    def stat_on_terminal_state(self, past_all_states, past_visit_count, past_visit_count_sum):
        """
        TODO: move to LIBEARTHQUAKE.SO
        """
        if past_visit_count == 0:
            banner = 'TERMINAL STATE(FRONTIER)'
            new_all_states = past_all_states + 1
        else:
            banner = 'TERMINAL STATE(REVISITED)'
            new_all_states = past_all_states
        LOG.info(
            colorama.Back.RED + '%s state %s, count=%d->%d, count_sum=%d->%d, all_states=%d->%d' + colorama.Style.RESET_ALL,
            banner,
            self.state.to_short_str(),
            past_visit_count, past_visit_count + 1,
            past_visit_count_sum, past_visit_count_sum + 1,
            past_all_states, new_all_states)

    def regist_state_to_libeq(self):
        json_dict = self.state.to_jsondict()
        json_str = json.dumps(json_dict)
        short_str = self.state.to_short_str()
        rc = self.oc.libearthquake.EQRegistExecutionHistory_UnstableAPI(short_str, json_str)
        assert rc == 0

    def on_terminal_state(self):
        LOG.debug(colorama.Back.RED +
                  '*** REACH TERMINAL STATE (%s) ***' +
                  colorama.Style.RESET_ALL, self.state.to_short_str())

        self.regist_state_to_libeq()

        ## make stat (TODO: move to LIBEARTHQUAKE.SO)
        all_states = len(self.visited_terminal_states)
        visit_count_sum = sum(self.visited_terminal_states.values())
        if self.state in self.visited_terminal_states:
            visit_count = self.visited_terminal_states[self.state]
        else:
            visit_count = 0
            self.visited_terminal_states[self.state] = 0
        self.stat_on_terminal_state(all_states, visit_count, visit_count_sum)
        self.visited_terminal_states[self.state] += 1

        ## notify termination to watchers
        for w in self.oc.watchers: w.on_terminal_state(self.state)

        ## Reset
        next_state = self.initial_state.make_copy()
        LOG.debug('Reset to %s', next_state.to_short_str())
        ## notify reset to watchers
        for w in self.oc.watchers: w.on_reset()
        return next_state


class DumbExplorer(ExplorerBase):
    def __init__(self, sleep_msecs=0):
        super(DumbExplorer, self).__init__()
        self.sleep_msecs = sleep_msecs

    def choose_digestible(self, digestibles):
        assert (digestibles)
        return digestibles[0]

    def call_action(self, action):
        if self.sleep_msecs:
            sleep(self.sleep_msecs / 1000.0)
        super(DumbExplorer, self).call_action(action)


class RandomExplorer(ExplorerBase):
    def __init__(self, time_slice):
        super(RandomExplorer, self).__init__()
        self.time_slice = time_slice  # msecs

    def choose_digestible(self, digestibles):
        assert (digestibles)
        r = random.randint(0, len(digestibles) - 1)
        chosen_digestible = digestibles[r]
        return chosen_digestible


class TimeBoundedRandomExplorer(RandomExplorer):
    def __init__(self, time_slice, time_bound):
        super(TimeBoundedRandomExplorer, self).__init__(time_slice)
        self.saved_time_slice = time_slice
        self.time_bound = time_bound  # msecs

    def choose_digestible(self, digestibles):
        assert (digestibles)
        now = time.time()
        hurried = filter(lambda d: (now - d.event.recv_timestamp) * 1000.0 > self.time_bound, digestibles)
        if len(hurried) > 0:
            LOG.debug('Hurried to send the following %d digestibles, now=%s', len(hurried), now)
            LOG.debug(hurried)
            self.time_slice = 0
            chosen_digestible = hurried[0]
        else:
            self.time_slice = self.saved_time_slice
            r = random.randint(0, len(digestibles) - 1)
            chosen_digestible = digestibles[r]
        return chosen_digestible


class GreedyExplorer(ExplorerBase):
    def __init__(self, time_slice):
        super(GreedyExplorer, self).__init__(time_slice)
        raise NotImplementedError(
            "GreedyExplorer is under refactoring since July 8, 2015. This will revive when new graph storage is implemented (Issue #23)")

    def choose_digestible(self, digestibles):
        pass

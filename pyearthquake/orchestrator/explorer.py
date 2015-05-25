## FIXME: remove unused imports
from abc import ABCMeta, abstractmethod
import colorama
import copy
import eventlet
from eventlet.greenthread import sleep
from eventlet.semaphore import Semaphore
from eventlet.timeout import Timeout
from eventlet.queue import *
from greenlet import GreenletExit
import networkx as nx
import six
import time
import random
import uuid
import json

from .. import LOG as _LOG
from ..entity import *
from ..util import *

from .state import *
from .watcher import *
from .digestible import *

LOG = _LOG.getChild('orchestrator.explorer')

class Graph(object):
    """
    MOVE ME TO LIBEARTHQUAKE.SO
    """
    def __init__(self, initial_state):
        self._g = nx.DiGraph()
        self.visit_node(initial_state)

    def draw(self):
        nx.draw(self._g)
        matplotlib.peclot.show()
        
    def get_leaf_nodes(self):
        return [n for n,d in self._g.out_degree().items() if d==0]

    def _print_nodes(self):
        leaf_nodes = self.get_leaf_nodes()
        LOG.debug('* Nodes (%d): %s', len(self._g.nodes()), [str(x) for x in self._g.nodes()])
        LOG.debug('* Leaf Nodes (%d): %s', len(leaf_nodes), [str(x) for x in leaf_nodes])

    def visit_node(self, state):
        assert isinstance(state, StateBase)
        count = self._g.node[state]['count'] if self._g.has_node(state) else 0
        # LOG.debug('Visit state %s, count=%d->%d', state.to_short_str(), count, count+1)
        self._g.add_node(state, count=count+1)
        
    def visit_edge(self, state, next_state, digestible):
        assert isinstance(state, StateBase)
        assert isinstance(next_state, StateBase)
        assert isinstance(digestible, DigestibleBase)
        self.visit_node(state)
        self.visit_node(next_state)
        self._g.add_edge(state, next_state, digestible=digestible)
        # self._print_nodes()


@six.add_metaclass(ABCMeta)
class ExplorerBase(object):
    def __init__(self):
        self.graph = None
        self._event_q = Queue()
        self.oc = None
        self.state = None
        self.initial_state = None
        self.visited_terminal_states = {} #key: state, value: count (TODO: MOVE TO LIBEARTHQUAKE.SO)
        self.time_slice = 0

    def init_with_orchestrator(self, oc, initial_state):
        self.oc = oc
        self.initial_state = initial_state
        self.state = self.initial_state.make_copy()
        LOG.debug(colorama.Back.BLUE +
                  'set initial state=%s' +
                  colorama.Style.RESET_ALL, self.state.to_short_str())
        self.graph = Graph(self.state)


    def send_event(self, event):
        assert isinstance(event, EventBase)
        self._event_q.put(event)
        
    def recv_events(self, timeout_msecs):
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


    def _worker__print_events_and_callbacks(self, digestibles, new_events, new_digestibles):
        if digestibles:
            LOG.debug('Before state %s, the following OLD %d callbacks had been yielded', self.state.to_short_str(), len(digestibles))
            for digestible in digestibles: LOG.debug('* %s', digestible)
        LOG.debug('In state %s, the following %d events happend', self.state.to_short_str(), len(new_events))
        for e in new_events:
            try: LOG.debug('* %f: %s', e.recv_timestamp, e.abstract_msg)
            except Exception: LOG.debug('* %s', e)
        LOG.debug('In state %s, the following NEW %d callbacks were yielded for the above %d events', self.state.to_short_str(), len(new_digestibles), len(new_events))
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
            self._worker__print_events_and_callbacks(digestibles, new_events, new_digestibles)
            digestibles.extend(new_digestibles)
            if not digestibles: LOG.warn('No DIGESTIBLE, THIS MIGHT CAUSE FALSE DEADLOCK, state=%s', self.state.to_short_str())

            next_state, digestibles = self.do_it(digestibles)
            if not digestibles: LOG.warn('No DIGESTIBLE, THIS MIGHT CAUSE FALSE DEADLOCK, next_state=%s', next_state.to_short_str())

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
        assert(any(digestible.event.uuid == chosen_digestible.event.uuid for digestible in digestibles))
        digestibles.remove(chosen_digestible)
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

    def do_transition(self, digestible):
        assert isinstance(digestible, DigestibleBase)
        LOG.debug(colorama.Back.BLUE +
                  'Invoking the callback at state=%s, digestible=%s' +
                  colorama.Style.RESET_ALL, self.state.to_short_str(), digestible)
        self.oc.call_action(digestible.action)
        next_state = self.state.make_copy()
        next_state.append_digestible(digestible)
        LOG.debug(colorama.Back.BLUE +
                  'State Transition: %s->%s' +
                  colorama.Style.RESET_ALL, self.state.to_short_str(), next_state.to_short_str())        

        self.graph.visit_edge(self.state, next_state, digestible)
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
        LOG.info(colorama.Back.RED + '%s state %s, count=%d->%d, count_sum=%d->%d, all_states=%d->%d' + colorama.Style.RESET_ALL,
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
        


class RandomExplorer(ExplorerBase):
    def __init__(self, time_slice):
        super(RandomExplorer, self).__init__()
        self.time_slice = time_slice
        
    def choose_digestible(self, digestibles):
        assert (digestibles)
        r = random.randint(0, len(digestibles)-1)
        chosen_digestible = digestibles[r]
        return chosen_digestible


class DumbExplorer(ExplorerBase):        
    def choose_digestible(self, digestibles):
        assert (digestibles)
        return digestibles[0]


from networkx.algorithms.traversal.depth_first_search import dfs_tree
class GreedyExplorer(ExplorerBase):
    def __init__(self, time_slice):
        super(GreedyExplorer, self).__init__()
        self.time_slice = time_slice
    
    def get_subtrees(self, digestibles):
        d = {}        
        frontier_digestibles = list(digestibles) # this is a shallow copy
        g = self.graph._g ## FIXME: should not access others' private vars  
        assert self.state in g.edge        
        for next_state in g.edge[self.state]:
            ## NOTE: even if digestible==edge_digestible, event_uuid can differ. Do NOT return edge_digestible.
            edge_digestible = g.edge[self.state][next_state]['digestible']
            digestibles_matched = [digestible for digestible in digestibles if digestible == edge_digestible]
            if not digestibles_matched: continue
            digestible = digestibles_matched[0]
            frontier_digestibles.remove(digestible)
            subtree = dfs_tree(g, next_state)
            d[digestible] = subtree
        for digestible in frontier_digestibles:
            d[digestible] = None
        return d

    def evaluate_digestible_subtree(self, digestible, subtree):
        assert(digestible) # subtree may be None
        if not subtree: 
            metric = 1.0
        else:
            subtree_nodes = subtree.number_of_nodes()
            metric = 1.0 / subtree_nodes if subtree_nodes > 0 else 1.0
        rand_factor = random.randint(9, 11) / 10.0
        metric *= rand_factor
        return metric
    
    def choose_digestible(self, digestibles):
        assert (digestibles)        
        digestible_metrics = {}
        for digestible, subtree in self.get_subtrees(digestibles).items():
            metric = self.evaluate_digestible_subtree(digestible, subtree)
            LOG.debug('Evaluated: metric=%f, digestible=%s', metric, digestible)
            digestible_metrics[digestible] = metric
        chosen_digestible = max(digestible_metrics, key=digestible_metrics.get)
        return chosen_digestible

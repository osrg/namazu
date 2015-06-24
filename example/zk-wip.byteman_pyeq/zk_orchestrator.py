#!/usr/bin/env python

from pyearthquake.orchestrator.state import *
from pyearthquake.orchestrator.orchestrator import *

class ZkState(BasicState):
    pass


class ZkOrchestrator(BasicOrchestrator):
    def make_initial_state(self):
        return ZkState()

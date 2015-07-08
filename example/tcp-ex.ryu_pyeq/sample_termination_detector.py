#!/usr/bin/env python

from pyearthquake.orchestrator.state import *
from pyearthquake.orchestrator.detector import *


class SampleTerminationDetector(TerminationDetectorBase):
    def __init__(self, messages=2):
        self.messages = messages
    
    def is_terminal_state(self, state):
        if state.forcibly_terminated:
            return True
        # terrible bad hack
        state_str = state.to_short_str()
        return all(str('(%d)' % msg_no) in state_str for msg_no in range(self.messages))

#!/usr/bin/env python

from pyearthquake.entity.event import *
from pyearthquake.orchestrator.state import *
from pyearthquake.orchestrator.orchestrator import *

class SampleState(BasicState):
    def to_short_str(self):
        """
        human-readable representation
        e.g. '<S 10(1)(0)>' for sequence Req1->Req0->Res1->Res0.
        """
        try:
            s = ''
            for d in self.digestible_sequence:
                if isinstance(d.event, PacketEvent):
                    msg = d.event.option['message']
                    msg_no = msg['msg_no']
                    is_res = msg['is_res']
                    s += ('(%d)' if is_res else '%d') % msg_no
                else:
                    s += '?[%s]' % d
            return '<S %s>' % s
        except Exception as e:
            LOG.exception(e)
            s = super(SampleState, self).to_short_str()
            return '<Bad S: %s>' % s
        

class SampleOrchestrator(BasicOrchestrator):
    def make_initial_state(self):
        return SampleState()

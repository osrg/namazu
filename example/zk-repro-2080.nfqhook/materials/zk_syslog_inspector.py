#!/usr/bin/env python
import pyearthquake
from pyearthquake.inspector.syslog import BasicSyslogInspector

LOG = pyearthquake.LOG.getChild(__name__)

class Zk2080SyslogInspector(BasicSyslogInspector):
    pass

if __name__ == '__main__':
    d = Zk2080SyslogInspector()
    d.start()

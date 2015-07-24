import pyearthquake
LOG = pyearthquake.LOG.getChild(__name__)

import sys
import json
from .common import *

def load_additional_modules(config):
    module_strs = []    
    try:
        module_strs = config['globalFlags']['plugin']['modules']
    except KeyError as e:
        return
    for f in module_strs:
        LOG.info('Loading module %s', f)
        __import__(f)

def load_orchestrator_plugin(config):
    oc_str = config['globalFlags']['plugin']['orchestrator']
    LOG.info('Loading orchestrator "%s"', oc_str)
    oc = eval(oc_str)
    assert isinstance(oc, pyearthquake.orchestrator.orchestrator.OrchestratorBase)
    LOG.info('Loaded orchestrator %s', oc)    
    return oc
    
def main():
    assert len(sys.argv) > 1, "argument is required (config path)"
    config = load_config(sys.argv[1])
    load_modules()
    load_additional_modules(config)    
    oc = load_orchestrator_plugin(config)
    oc.init_with_config(config)
    LOG.info('Starting orchestrator')
    oc.start()

if __name__ == "__main__":
    main()

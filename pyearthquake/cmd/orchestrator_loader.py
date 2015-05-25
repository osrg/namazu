import pyearthquake
LOG = pyearthquake.LOG.getChild('cmd.orchestrator_loader')

import sys
import json

def load_modules():
    import pyearthquake
    import pyearthquake.orchestrator.orchestrator
    import pyearthquake.orchestrator.watcher
    import pyearthquake.orchestrator.detector
    import pyearthquake.orchestrator.explorer

def load_config():
    config_path = sys.argv[1]
    config_file = open(config_path)
    config_str = config_file.read()
    config_file.close()
    config = json.loads(config_str)
    return config

def load_orchestrator_plugin(config):
    oc_str = config['globalFlags']['plugin']['orchestrator']
    LOG.info('Loading orchestrator "%s"', oc_str)
    oc = eval(oc_str)
    assert isinstance(oc, pyearthquake.orchestrator.orchestrator.OrchestratorBase)
    LOG.info('Loaded orchestrator %s', oc)    
    return oc
    
def main():
    assert len(sys.argv) > 1, "argument is required (config path)"
    load_modules()
    config = load_config()
    oc = load_orchestrator_plugin(config)
    oc.init_with_config(config)
    LOG.info('Starting orchestrator')
    oc.start()

if __name__ == "__main__":
    main()

import pyearthquake
LOG = pyearthquake.LOG.getChild(__name__)

import json

def load_modules():
    import sys
    import pyearthquake
    import pyearthquake.orchestrator.orchestrator

def parse_config(config_str):
    # TODO: add support for YAML
    config = json.loads(config_str)
    return config

def load_config(config_path):
    config_file = open(config_path)
    config_str = config_file.read()
    config_file.close()
    config = parse_config(config_str)
    return config

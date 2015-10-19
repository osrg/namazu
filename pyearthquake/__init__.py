from . import common
LOG = common.init_logger()

# Eventlet misusage detection (Do NOT use this with a non-eventlet blocking function)
# import eventlet.debug
# LOG.debug('eventlet.debug.hub_blocking_detection(True, resolution=3)')
# eventlet.debug.hub_blocking_detection(True, resolution=10)

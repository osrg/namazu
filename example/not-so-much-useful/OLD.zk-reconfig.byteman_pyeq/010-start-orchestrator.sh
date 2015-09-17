#! /bin/bash
set -x

CONFIG_JSON=config.json

EQDIR=$(pwd)/../..
export LD_LIBRARY_PATH=$EQDIR/bin:$LD_LIBRARY_PATH
export PYTHONPATH=$EQDIR:$PYTHONPATH
python -m pyearthquake.cmd.orchestrator_loader $CONFIG_JSON

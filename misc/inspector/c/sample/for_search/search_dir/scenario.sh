#! /bin/bash

NMZ_ENV_PROCESS_ID="node1" ../cmd &
NMZ_ENV_PROCESS_ID="node2" ../cmd &

sleep 2

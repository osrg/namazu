#! /bin/bash

EQ_ENV_PROCESS_ID="node1" ../cmd &
EQ_ENV_PROCESS_ID="node2" ../cmd &

sleep 2

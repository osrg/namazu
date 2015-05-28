#!/bin/bash

set -x

for f in $(seq 1 3); do \
        docker exec -d zk${f} /bin/bash -c '/init.py > /log 2>&1'; \
      done

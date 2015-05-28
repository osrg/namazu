#!/bin/bash

set -x

for f in $(seq 1 3); do \
        docker run -i -t -d -e ZKID=${f} -h zk${f} --name zk${f} zk_testbed /bin/bash; \
      done

#!/bin/bash

set -x

for f in $(seq 1 3); do \
	# TODO: add docs about ovsbr0, 192.168.42.0/24
	# TODO: make them configurable
        sudo pipework ovsbr0 zk${f} 192.168.42.${f}/24; \
      done

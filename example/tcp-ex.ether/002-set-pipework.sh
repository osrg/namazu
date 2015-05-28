#!/bin/bash

set -x

sudo pipework ovsbr0 tcp-ex-server 192.168.42.1/24
sudo pipework ovsbr0 tcp-ex-client 192.168.42.2/24

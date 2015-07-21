#!/bin/bash

set -x

docker exec  tcp-ex-client /tcp-ex/tcp-ex -client -peer 192.168.42.1:9999 -workers 2 -messages 2 

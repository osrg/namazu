#!/bin/sh
set -x

killall -9 java     #FIXME

rm -rf /tmp/eq/zk*   #NOTE: you must keep /tmp/eq/search

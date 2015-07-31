#!/bin/bash
#set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

grep "BUILD SUCCESSFUL" ${EQ_WORKING_DIR}/zk-test.log
exit $?

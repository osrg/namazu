#!/bin/bash
#set -e # exit on an error
. ${NMZ_MATERIALS_DIR}/lib.sh

grep "BUILD SUCCESSFUL" ${NMZ_WORKING_DIR}/zk-test.log
exit $?

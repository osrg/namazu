#!/bin/bash

set -e # exit on an error
. ${EQ_MATERIALS_DIR}/lib.sh

CHECK_PREREQUISITES
FETCH_JACOCO
FETCH_ZK
BUILD_ZK

exit 0

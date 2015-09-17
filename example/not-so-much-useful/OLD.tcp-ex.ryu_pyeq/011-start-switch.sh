#!/bin/bash
set -x
export PYTHONPATH=../..
ryu-manager --verbose ./sample_switch.py 

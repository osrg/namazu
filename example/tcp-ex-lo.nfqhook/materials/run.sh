#!/bin/bash
set -e # exit on an error
. $EQ_MATERIALS_DIR/lib.sh

export EQ_ETHER_ZMQ_ADDR="ipc://${EQ_WORKING_DIR}/ether_inspector"


# PAUSE

IMPORTANT "Please make sure you have set up iptables"
IMPORTANT "Please make sure you have set PYTHONPATH(=${PYTHONPATH})"
INFO "Starting NFQHook (ZMQ: ${EQ_ETHER_ZMQ_ADDR})"
python $EQ_MATERIALS_DIR/sample_nfqhook.py 2>&1 &
NFQHOOK_PID=$!
INFO "PID: $NFQHOOK_PID"

INFO "Starting Inspector (ZMQ: ${EQ_ETHER_ZMQ_ADDR})"
python $EQ_MATERIALS_DIR/sample_inspector.py 2>&1 &
INSPECTOR_PID=$!
INFO "PID: $INSPECTOR_PID"

SLEEP 3

IMPORTANT "Please make sure you have built tcp-ex binary"
IMPORTANT "Please make sure you have made a user \"nfqhooked\""

INFO "Starting tcp-ex server"
chmod -R 777 ${EQ_WORKING_DIR} # FIXME: 777 for nfqhooked
sudo -E -u nfqhooked sh -c "echo \$\$ > ${EQ_WORKING_DIR}/server.pid; exec $EQ_MATERIALS_DIR/tcp-ex/tcp-ex -server"  2>&1 &
SLEEP 3 # wait for sever.pid file
SERVER_PID=$(cat ${EQ_WORKING_DIR}/server.pid)
INFO "PID: $SERVER_PID"

INFO "Starting tcp-ex client"
sudo -u nfqhooked $EQ_MATERIALS_DIR/tcp-ex/tcp-ex -client -messages 2 -workers 2

INFO "tcp-ex client finished"
INFO "killing $NFQHOOK_PID $INSPECTOR_PID $SERVER_PID"
kill -9 $NFQHOOK_PID $INSPECTOR_PID $SERVER_PID

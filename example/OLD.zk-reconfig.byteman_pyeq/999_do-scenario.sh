#!/bin/bash

func_do_scenario(){
    rm -rf /tmp/eq/*
    pkill -9 java
    pkill -9 python

    if [ ! -e /tmp/eq_data ]; then
        mkdir /tmp/eq_data
    fi

    TARGET=/tmp/eq_data/$(date +"%Y%m%d.%H%M%S")
    mkdir ${TARGET}
    CONSOLE_LOG=${TARGET}/console.log

    bash 000-prepare-zk.sh >> ${CONSOLE_LOG} 2>&1
    sleep 1
    bash 010-start-orchestrator.sh >> ${CONSOLE_LOG} 2>&1 &
    sleep 3
    bash 020-start-zk-ensemble.sh >> ${CONSOLE_LOG} 2>&1
    sleep 3
    bash 030-leader-change-background-concurrent-write.sh >> ${CONSOLE_LOG} 2>&1
    sleep 1
    bash 031-check-zknode.sh >> ${CONSOLE_LOG} 2>&1
    sleep 1
    bash 040-inspection-end.sh >> ${CONSOLE_LOG} 2>&1
    sleep 1
    bash 041-kill-zk-ensemble.sh >> ${CONSOLE_LOG} 2>&1
    sleep 1
    pkill -9 python >>${CONSOLE_LOG} 2>&1

    cp /tmp/eq/search/history/*/{json,name} ${TARGET} >> ${CONSOLE_LOG} 2>&1
    sleep 10
}


if [ $# -ne 1 ]; then
  echo "error"
  exit 1
fi

for i in `seq 1 ${1}`;
do
    func_do_scenario
done



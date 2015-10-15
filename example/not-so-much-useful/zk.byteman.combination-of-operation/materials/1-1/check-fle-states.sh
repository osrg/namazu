#! /bin/bash
set -x

result=1
count=0
while [ ${result} -eq 1 -a ${count} -lt 30 ]
do
    python ${EQ_MATERIALS_DIR}/check-fle-states.py $@ && result=0
    (( count++ ))
    sleep 1
done

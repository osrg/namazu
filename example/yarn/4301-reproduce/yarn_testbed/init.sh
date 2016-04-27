#!/bin/bash
set -e # exit on an error

function INFO(){
    echo -e "\e[104m\e[97m[INFO]\e[49m\e[39m $@"
}

INFO "Checking wheter /nmzfs is mounted"
ls /nmzfs > /dev/null

INFO "Starting SSH"
service ssh start

INFO "Starting YARN"
start-yarn.sh

INFO "Please open http://localhost:8042 and check NodeHealthyStatus"

INFO "^C to exit.."
while true; do sleep 5; done

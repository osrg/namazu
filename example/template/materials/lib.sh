#!/bin/bash

function INFO(){
    echo -e "\e[104m\e[97m[INFO]\e[49m\e[39m $@"
}

function IMPORTANT(){
    echo -e "\e[105m\e[97m[IMPORTANT]\e[49m\e[39m $@"
}

function SLEEP(){
    echo -n $(INFO "Sleeping(${1} secs)..")
    sleep ${1}
    echo "Done"
}

function PAUSE(){
    TMP=$(mktemp /tmp/namazu-wait.XXXXX)
    IMPORTANT "PAUSING. remove ${TMP} to continue"
    while [ -e $TMP ]; do
      sleep 1
    done
}


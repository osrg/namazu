#!/bin/bash
echo "${0}: entered"
TMP=$(mktemp)
echo "${0}: remove ${TMP} to kill me"
while [ -e $TMP ]; do
    sleep 3
done
echo "${0}: leaving"

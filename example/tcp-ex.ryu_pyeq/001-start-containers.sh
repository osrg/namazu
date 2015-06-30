#!/bin/bash

set -x

docker run -i -t -d --name tcp-ex-server -h tcp-ex-server -v $(pwd)/tcp-ex:/tcp-ex ubuntu:latest /tcp-ex/tcp-ex -server
docker run -i -t -d --name tcp-ex-client -h tcp-ex-client -v $(pwd)/tcp-ex:/tcp-ex ubuntu:latest /bin/bash

#! /bin/bash

protoc --java_out=src/ -I../../ ../../i2g_message.proto

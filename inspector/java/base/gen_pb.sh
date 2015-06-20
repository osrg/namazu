#! /bin/bash

protoc --java_out=src/ -I../../ ../../inspector_message.proto

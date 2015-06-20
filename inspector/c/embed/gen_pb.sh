#! /bin/bash

protoc --cpp_out=. -I../../ ../../inspector_message.proto

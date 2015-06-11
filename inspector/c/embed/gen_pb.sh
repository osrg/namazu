#! /bin/bash

protoc --cpp_out=. -I../../ ../../i2g_message.proto

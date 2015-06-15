#! /bin/bash

protoc --go_out=. -I../../inspector ../../inspector/i2g_message.proto
protoc --go_out=. -I. ./o2g_message.proto


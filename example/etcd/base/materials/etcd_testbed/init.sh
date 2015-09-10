#! /bin/bash

etcd -name etcd$ETCDID -listen-client-urls http://192.168.42.$ETCDID:4001 -advertise-client-urls http://192.168.42.$ETCDID:4001 -listen-peer-urls http://192.168.42.$ETCDID:7001 -initial-advertise-peer-urls http://192.168.42.$ETCDID:7001 -initial-cluster-token etcd-cluster-1 -initial-cluster 'etcd1=http://192.168.42.1:7001,etcd2=http://192.168.42.2:7001,etcd3=http://192.168.42.3:7001' -initial-cluster-state new -heartbeat-interval=100 -election-timeout=500

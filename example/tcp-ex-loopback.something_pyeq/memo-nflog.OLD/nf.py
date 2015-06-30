#!/usr/bin/python

# need root privileges

import struct
import sys
import time

from socket import AF_INET, AF_INET6, inet_ntoa

import nflog

from dpkt import ip

def cb(payload):
	print "python callback called !"

	print "payload len ", payload.get_length()
        print "nfmark ", payload.get_nfmark()
	data = payload.get_data()
	pkt = ip.IP(data)
	print "proto:", pkt.p
	print "source: %s" % inet_ntoa(pkt.src)
	print "dest: %s" % inet_ntoa(pkt.dst)
	if pkt.p == ip.IP_PROTO_TCP:
	 	print "  sport: %s" % pkt.tcp.sport
	 	print "  dport: %s" % pkt.tcp.dport

	sys.stdout.flush()
	return 1

l = nflog.log()

print "setting callback"
l.set_callback(cb)

print "open"
l.fast_open(42, AF_INET)

print "trying to run"
try:
	l.try_run()
except KeyboardInterrupt, e:
	print "interrupted"


print "unbind"
l.unbind(AF_INET)

print "close"
l.close()


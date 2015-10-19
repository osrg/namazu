#!/usr/bin/python
## FIXME: move these ones to the config file
PORTS = []

import telnetlib
import re
import sys

def parse_stat_str(s):
    print 'Parsing: %s' % s
    d = {'mode': 'NOTPARSED', 'zxid': 'NOTPARSED'}
    d['mode'] = re.compile('.*Mode:\s(.*)').search(s).group(1)
    d['zxid'] = re.compile('.*Zxid:\s(.*)').search(s).group(1)
    print 'Parsed: %s' % d
    return d


def stat(host='localhost', port='2181'):
    tn = telnetlib.Telnet(host, port)
    tn.write('stat')
    s = tn.read_all()
    try:
        d = parse_stat_str(s)
    except Exception as e:
        print 'Could not parse: %s' % s
        raise e
    tn.close()
    return d


def get_stats():
    l = []
    for port in PORTS:
        d = stat(port=port)
        l.append(d)
    return l


def main():
    l = get_stats()
    leaders = filter(lambda d: d['mode'] == 'leader', l)
    observers = filter(lambda d: d['mode'] == 'observer', l)
    assert len(leaders) == 1, 'Bad leader election: %s' % l
    assert len(observers) == 0, 'Bad observers: %s' % l
    print('OK (%d leaders, %d observers): %s' % (len(leaders), len(observers), l))
    

if __name__ == '__main__':
    port_list = sys.argv
    del port_list[0]
    for port in port_list:
        print(port)
        PORTS.append(port)

    main()

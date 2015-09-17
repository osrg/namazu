Earthquake Demo (Ethernet Inspector + nfqhook)
===
NFQHOOK does not require OVS+ryu+Docker+pipework.

Run
---
	
	$ cd ~/WORK/earthquake
	$ ( cd example/tcp-ex-lo.nfqhooked/materials/tcp-ex; go build)
	$ sudo useradd --foo-bar nfqhooked
	$ sudo iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num 42
	$ sudo rm -rf /tmp/foo; mkdir /tmp/foo
	$ bin/earthquake init example/tcp-ex-lo.nfqhook/config.yaml example/tcp-ex-lo.nfqhook/materials/ /tmp/foo
	$ sudo PYTHONPATH=$(pwd) bin/earthquake run /tmp/foo

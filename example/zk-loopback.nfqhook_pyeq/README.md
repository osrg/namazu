Usage
---
    
    $ useradd --some-options nfqhooked
    $ iptables -A OUTPUT -p tcp -m owner --uid-owner $(id -u nfqhooked) -j NFQUEUE --queue-num 42
    $ ln -s SOMEWHERE/zookeeper .
	$ sudo PYTONPATH=../.. ./zk_nfqhook.py
    $ ./000-prepare.sh.
    $ ./100-run-experiment.sh
    $ ./100-run-experiment.sh
    $ ./100-run-experiment.sh
    $ ./100-run-experiment.sh
    $ ./999-end.sh
    

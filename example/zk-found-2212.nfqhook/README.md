# ZooKeeper Bug [ZOOKEEPER-2212](https://issues.apache.org/jira/browse/ZOOKEEPER-2212): distributed race condition related to QV version

Identical to [zk-found-2212.ryu](../zk-found-2212.ryu), but does not depend on OVS+pipework+ryu+Docker.

    $ sudo pip install git+https://github.com/twitter/zktraffic@68d9f85d8508e01f5d2f6657666c04e444e6423c  #(Jul 18, 2015)
    $ sudo pip install hookswitch
    $ sudo sh -c 'echo 0 > /proc/sys/net/ipv4/tcp_autocorking' # recommended if you are using Linux 3.14 or later
    $ sudo useradd -m nfqhooked # this user is needed for internal sandboxing
    $ sudo PYTHONPATH=$(pwd)/../../misc ../../bin/nmz init --force config.toml materials /tmp/zk-2212
    $ sudo PYTHONPATH=$(pwd)/../../misc ../../bin/nmz run /tmp/zk-2212
    

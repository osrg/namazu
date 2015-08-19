# ZooKeeper Bug [ZOOKEEPER-2251](https://issues.apache.org/jira/browse/ZOOKEEPER-2251): Add Client side packet response timeout to avoid infinite wait.

Trying to reproduce, but not yet succeeded.


## How to Reproduce the Bug with Earthquake


### Start Earthquake
Please see [../../doc/how-to-setup-env.md](../../doc/how-to-setup-env.md) for how to setup the environment.

    $ cp ../../bin/earthquake-inspector.jar ./materials
    $ ../../bin/earthquake init --force config.toml materials /tmp/zk-2251


### Run Experiments

    $ ../../bin/earthquake run /tmp/zk-2251


### Environment Variables

* `EQ_DISABLE`(default: (unset)): disable the substantial part of Earthquake if set
* `ZK_GIT_COMMIT`(default:(see `materials/lib.sh`)) : use another ZooKeeper version
* `ZK_SOURCE_DIR`(default: (unset)) : use another ZooKeeper source directory if set

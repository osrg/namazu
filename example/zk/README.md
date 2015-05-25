# Earthquake Demo (zookeeper)

## Getting Started
Prepare

    $ sudo apt-get install -y \
      python-eventlet python-flask python-colorama python-networkx python-six \
      default-jdk maven
    $ cp config_example.json config.json
    $ ./000-prepare-zk.sh
    $ ./010-start-orchestrator.sh


Run experiments

    $ ./020-start-zk-ensemble.sh
    $ ./030-concurrent-write.sh
    $ ./040-inspection-end.json
    $ ./041-kill-zk-ensemble.sh

    $ ./020-start-zk-ensemble.sh
    $ ./030-concurrent-write.sh
    $ ./040-inspection-end.json
    $ ./041-kill-zk-ensemble.sh

    # loop as many times as you want...


Get experimental result CSV

    $ curl http://localhost:10000/visualize_api/csv
    # exp_count	patterns
    1 1
    5 2
    ..
    

## SEE ALSO
example-output/README.md    

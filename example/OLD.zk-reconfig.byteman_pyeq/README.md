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
    $ ./040-inspection-end.sh
    $ ./041-kill-zk-ensemble.sh

    $ ./020-start-zk-ensemble.sh
    $ ./030-concurrent-write.sh
    $ ./040-inspection-end.sh
    $ ./041-kill-zk-ensemble.sh

    # loop as many times as you want...


   ..

Get execution history

    $ jq . < /tmp/eq/search/history/0000000000000001/json
    {
      "elements": [
        {
          "process": "zksrv1",
          "action_digest": "AcceptDeferredEventAction",
          "event_digest": [
            "FunctionCallEvent",
            {
              "func_name": "FollowerRequestProcessor.processRequest",
              "request": "sessionid:0x10001facd6f0000 type:createSession cxid:0x0 zx
    id:0xfffffffffffffffe txntype:unknown reqpath:n/a"
            }
          ]
        },
        {
          "process": "zksrv2",
          "action_digest": "AcceptDeferredEventAction",
          "event_digest": [
            "FunctionCallEvent",
            {
              "func_name": "LeaderRequestProcessor.processRequest",
              "request": "sessionid:0x20001facbe30000 type:createSession cxid:0x0 zx
    id:0xfffffffffffffffe txntype:unknown reqpath:n/a"
            }
          ]
        },
        ..
    }
    


## SEE ALSO
example-output/README.md    

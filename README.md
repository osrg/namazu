# Earthquake: Dynamic Model Checker for Distributed Systems

## Python Binding
       
    $ ./build 
    # make sure ./bin/libearthquake.so is built
    $ LD_LIBRARY_PATH=./bin PYTHONPATH=. python -m pyearthquake.cmd.orchestrator_loader example/zk/config.json
    # (see example/zk/README.md for further information)
    
   

## Earthquake Classic (no binding)
   
    $ ./build
    $ ./bin/earthquake-classic -help



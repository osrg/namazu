# How to setup the environment for Earthquake Container (simplified CLI)
For full-stack environment, please refer to [how-to-setup-env-full.md](how-to-setup-env-full.md).


    $ sudo apt-get install libzmq3-dev libnetfilter-queue-dev
    $ go get github.com/osrg/earthquake/earthquake-container
    $ sudo earthquake-container run -it --rm --eq-config config.toml ubuntu bash

Example configuration file (`config.toml`):

```toml
explorePolicy = "random"
[explorePolicyParam]
  minInterval = "80ms"
  maxInterval = "3000ms"
  #shellActionInterval = "5000ms"
  #shellActionCommand = "echo hello $(date)"
  #faultActionProbability = 0.0
```

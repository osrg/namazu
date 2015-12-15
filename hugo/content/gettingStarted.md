+++
categories = ["general"]
date = "2015-12-15"
tags = ["document"]
title = "Getting Started"

+++
For full-stack environment, please refer to [how-to-setup-env-full.md](https://github.com/osrg/earthquake/blob/master/doc/how-to-setup-env-full.md).

    $ sudo apt-get install libzmq3-dev libnetfilter-queue-dev
    $ go get github.com/osrg/earthquake/earthquake-container
    $ sudo earthquake-container run -it --rm --eq-config config.toml ubuntu bash

Example configuration file (`config.toml`):


    explorePolicy = "random"
    [explorePolicyParam]
      minInterval = "80ms"
      maxInterval = "3000ms"
      #shellActionInterval = "5000ms"
      #shellActionCommand = "echo hello $(date)"
      #faultActionProbability = 0.0
    

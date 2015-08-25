## Dockerfile for Earthquake
## Available at Docker Hub: osrg/earthquake
FROM osrg/dind-ovs-ryu
MAINTAINER Akihiro Suda <suda.akihiro@lab.ntt.co.jp>

RUN apt-get update && apt-get install -y --no-install-recommends \
    ## Install Earthquake deps (protoc, JDK)
    protobuf-compiler default-jdk maven \
    ## Install useful stuffs
    sudo \
    ## (Optional) Install MongoDB storage
    mongodb \
    ## (Optional) Install pyearthquake deps
    python-flask python-scapy python-zmq \
    ## (Optional) Install pyearthquake nfqhook deps
    libnetfilter-queue1 python-prctl

## Install Go 1.5
RUN curl https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz | tar Cxz /usr/local
ENV PATH /usr/local/go/bin:$PATH

## (Optional) Install pyearthquake deps
RUN pip install hexdump

## (Optional) Install pyearthquake ryu deps
RUN pip uninstall -y ryu && pip install ryu==3.20.2
RUN curl https://raw.githubusercontent.com/jpetazzo/pipework/master/pipework -o /usr/local/bin/pipework
RUN chmod +x /usr/local/bin/pipework

## Copy Earthquake to /earthquake
ADD . /earthquake
WORKDIR /earthquake
RUN ( git submodule init && git submodule update )

## Build Earthquake
RUN ./build

## Silence dind logs
ENV LOG file

## Start init (does NOT enable DinD/OVS/Ryu by default)
ADD docker/eq-init.py /eq-init.py
CMD ["/eq-init.py"]

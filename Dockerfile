## Dockerfile for Namazu
## Available at Docker Hub: osrg/namazu

FROM billyteves/ubuntu-dind:16.04 AS dind-ovs-ryu
MAINTAINER Akihiro Suda <suda.akihiro@lab.ntt.co.jp>

# Install OVS
RUN apt-get update && apt-get install -y openvswitch-switch

# Install Python packages
RUN bash -c 'apt-get install -y python-{colorama,dev,eventlet,lxml,msgpack,netaddr,networkx,oslo.config,paramiko,pip,routes,six,webob}'

# Install Ryu
RUN pip install ryu

# Install pipework
RUN apt-get install -y arping
RUN wget --no-check-certificate --quiet https://raw.githubusercontent.com/jpetazzo/pipework/master/pipework -O /usr/local/bin/pipework
RUN chmod +x /usr/local/bin/pipework

# Install misc useful stuffs
RUN apt-get install -y less lv netcat telnet bash-completion vim byobu

# Install init
ADD ./misc/dind-ovs-ryu/init.dind-ovs-ryu.sh /init.dind-ovs-ryu.sh
RUN chmod +x /init.dind-ovs-ryu.sh
CMD ["wrapdocker", "/init.dind-ovs-ryu.sh"]

FROM dind-ovs-ryu
MAINTAINER Akihiro Suda <suda.akihiro@lab.ntt.co.jp>

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ## Install Namazu deps
    protobuf-compiler pkg-config libzmq3-dev libnetfilter-queue-dev \
    ## (Optional) Install Java inspector deps
    default-jdk maven \
    ## (Optional) Install useful stuffs
    sudo ant ant-optional \
    ## (Optional) Install MongoDB storage
    mongodb \
    ## (Optional) Install FUSE inspector deps
    fuse \
    ## (Optional) Install pynmz deps
    python-flask python-scapy python-zmq \
    ## (Optional) Install pynmz nfqhook deps
    libnetfilter-queue1 python-prctl

## Install Go
RUN curl https://storage.googleapis.com/golang/go1.10.linux-amd64.tar.gz | tar Cxz /usr/local && mkdir /gopath
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /gopath

## (Optional) Install pynmz deps
RUN pip install hexdump requests

## (Optional) Install hookswitch
RUN pip install hookswitch==0.0.2

## (Optional) Install pipework for DinD
RUN curl https://raw.githubusercontent.com/jpetazzo/pipework/master/pipework -o /usr/local/bin/pipework
RUN chmod +x /usr/local/bin/pipework

## (Optional) Create a user for nfqueue sandbox
RUN useradd -m nfqhooked

## Copy Namazu to $GOPATH/src/github.com/osrg/namazu
RUN mkdir -p $GOPATH/src/github.com/osrg
ADD . $GOPATH/src/github.com/osrg/namazu
WORKDIR $GOPATH/src/github.com/osrg/namazu
RUN ( git submodule init && git submodule update )
ENV PYTHONPATH $GOPATH/src/github.com/osrg/namazu:$PYTHONPATH

## Build Namazu
RUN ./build

## Silence dind logs
ENV LOG file

## Start init (does NOT enable DinD/OVS/Ryu by default)
ADD misc/docker/nmz-init.py /nmz-init.py
CMD ["/nmz-init.py"]

# Based on akihirosuda/zookeeper-dynamic
FROM ubuntu:15.04
MAINTAINER mitake

ADD etcd /
ADD init.sh /

CMD ["bash", "init.sh"]

EXPOSE 4001 7001

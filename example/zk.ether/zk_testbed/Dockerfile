FROM akihirosuda/zookeeper-dynamic
MAINTAINER akihirosuda

ADD log4j.properties /zk/conf/
# docker-enter fails if space in env var
ENV ZOO_LOG4J_PROP INFO,CONSOLE,SYSLOG


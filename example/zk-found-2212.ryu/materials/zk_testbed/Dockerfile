# Based on akihirosuda/zookeeper-dynamic
FROM java:7
MAINTAINER akihirosuda

ENV ZOO_LOG4J_PROP DEBUG,CONSOLE,SYSLOG
ENV JAVA_TOOL_OPTIONS -Dfile.encoding=UTF8

RUN apt-get update && apt-get install -y ant telnet netcat less lv

RUN mkdir /jacoco && \
    cd /jacoco && \
    curl -L -O http://search.maven.org/remotecontent?filepath=org/jacoco/jacoco/0.7.5.201505241946/jacoco-0.7.5.201505241946.zip && \
    unzip jacoco-0.7.5.201505241946.zip

RUN mkdir /zk /zk_data
ADD zookeeper /zk
ADD init.py /
ADD log4j.properties /zk/conf/
WORKDIR /zk

RUN ant

ENV JVMFLAGS -javaagent:/jacoco/lib/jacocoagent.jar
CMD ["python", "/init.py"]
EXPOSE 2181 2888 3888

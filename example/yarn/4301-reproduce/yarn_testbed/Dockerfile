FROM akihirosuda/hadoop-docker-nightly:20151027
MAINTAINER Akihiro Suda

ADD yarn-site.xml /hadoop/etc/hadoop/

ADD init.sh /
RUN chmod +x /init.sh
CMD ["/init.sh"]
EXPOSE 8030 8031 8032 8033 8040 8042 8088

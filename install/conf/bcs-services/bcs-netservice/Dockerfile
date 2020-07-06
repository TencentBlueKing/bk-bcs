FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-netservice

ADD bcs-netservice /data/bcs/bcs-netservice/
ADD container-start.sh /data/bcs/bcs-netservice/
RUN chmod +x /data/bcs/bcs-netservice/container-start.sh

WORKDIR /data/bcs/bcs-netservice/
ENTRYPOINT ["/data/bcs/bcs-netservice/container-start.sh"]


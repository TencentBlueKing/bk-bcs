FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-gw-controller

ADD bcs-gw-controller /data/bcs/bcs-gw-controller/
ADD container-start.sh /data/bcs/bcs-gw-controller/
RUN chmod +x /data/bcs/bcs-gw-controller/container-start.sh

WORKDIR /data/bcs/bcs-gw-controller/
ENTRYPOINT ["/data/bcs/bcs-gw-controller/container-start.sh"]


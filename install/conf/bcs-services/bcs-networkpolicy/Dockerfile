FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-networkpolicy

ADD bcs-networkpolicy /data/bcs/bcs-networkpolicy/
ADD container-start.sh /data/bcs/bcs-networkpolicy/
RUN chmod +x /data/bcs/bcs-networkpolicy/container-start.sh

WORKDIR /data/bcs/bcs-networkpolicy/
ENTRYPOINT ["/data/bcs/bcs-networkpolicy/container-start.sh"]


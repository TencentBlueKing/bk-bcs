FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-gateway-discovery

ADD bcs-gateway-discovery /data/bcs/bcs-gateway-discovery/
ADD container-start.sh /data/bcs/bcs-gateway-discovery/
RUN chmod +x /data/bcs/bcs-gateway-discovery/container-start.sh

WORKDIR /data/bcs/bcs-gateway-discovery/
ENTRYPOINT ["/data/bcs/bcs-gateway-discovery/container-start.sh"]


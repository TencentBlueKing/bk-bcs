FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-network-detection

ADD bcs-network-detection /data/bcs/bcs-network-detection/
ADD container-start.sh /data/bcs/bcs-network-detection/
RUN chmod +x /data/bcs/bcs-network-detection/container-start.sh

WORKDIR /data/bcs/bcs-network-detection/
ENTRYPOINT ["/data/bcs/bcs-network-detection/container-start.sh"]


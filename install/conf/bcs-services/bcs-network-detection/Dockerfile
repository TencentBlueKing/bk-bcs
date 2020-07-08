FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-network-detection

ADD bcs-network-detection /data/bcs/bcs-network-detection/
ADD bcs-network-detection.json.template /data/bcs/bcs-network-detection/
ADD container-start.sh /data/bcs/bcs-network-detection/
RUN chmod +x /data/bcs/bcs-network-detection/container-start.sh

WORKDIR /data/bcs/bcs-network-detection/
CMD ["/data/bcs/bcs-network-detection/container-start.sh"]


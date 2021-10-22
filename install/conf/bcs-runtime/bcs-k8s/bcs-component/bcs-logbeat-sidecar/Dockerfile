FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-logbeat-sidecar

ADD bcs-logbeat-sidecar /data/bcs/bcs-logbeat-sidecar/
ADD bcs-logbeat-sidecar.json.template /data/bcs/bcs-logbeat-sidecar/
ADD container-start.sh /data/bcs/bcs-logbeat-sidecar/
RUN chmod +x /data/bcs/bcs-logbeat-sidecar/container-start.sh

WORKDIR /data/bcs/bcs-logbeat-sidecar/
CMD ["/data/bcs/bcs-logbeat-sidecar/container-start.sh"]

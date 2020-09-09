FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-mesh-manager

ADD bcs-mesh-manager /data/bcs/bcs-mesh-manager/
ADD bcs-mesh-manager.json.template /data/bcs/bcs-mesh-manager/
ADD container-start.sh /data/bcs/bcs-mesh-manager/
RUN chmod +x /data/bcs/bcs-mesh-manager/container-start.sh

WORKDIR /data/bcs/bcs-mesh-manager/
CMD ["/data/bcs/bcs-mesh-manager/container-start.sh"]

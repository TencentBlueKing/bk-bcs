FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-log-manager

ADD bcs-log-manager /data/bcs/bcs-log-manager/
ADD bcs-log-manager.json.template /data/bcs/bcs-log-manager/
ADD container-start.sh /data/bcs/bcs-log-manager/
RUN chmod +x /data/bcs/bcs-log-manager/container-start.sh

WORKDIR /data/bcs/bcs-log-manager/
CMD ["/data/bcs/bcs-log-manager/container-start.sh"]

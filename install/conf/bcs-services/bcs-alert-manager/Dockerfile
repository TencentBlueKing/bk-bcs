FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-alert-manager

ADD ./swagger/ /data/bcs/swagger
ADD bcs-alert-manager /data/bcs/bcs-alert-manager/
ADD container-start.sh /data/bcs/bcs-alert-manager/
ADD bcs-alert-manager.json.template /data/bcs/bcs-alert-manager/
RUN chmod +x /data/bcs/bcs-alert-manager/container-start.sh

WORKDIR /data/bcs/bcs-alert-manager/
CMD ["/data/bcs/bcs-alert-manager/container-start.sh"]

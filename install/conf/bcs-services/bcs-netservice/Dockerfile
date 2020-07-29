FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-netservice

ADD bcs-netservice /data/bcs/bcs-netservice/
ADD bcs-netservice.json.template /data/bcs/bcs-netservice/
ADD container-start.sh /data/bcs/bcs-netservice/
RUN chmod +x /data/bcs/bcs-netservice/container-start.sh

WORKDIR /data/bcs/bcs-netservice/
CMD ["/data/bcs/bcs-netservice/container-start.sh"]


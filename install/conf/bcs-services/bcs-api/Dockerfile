FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-api

ADD bcs-api /data/bcs/bcs-api/
ADD bcs-api.json.template /data/bcs/bcs-api/
ADD container-start.sh /data/bcs/bcs-api/
RUN chmod +x /data/bcs/bcs-api/container-start.sh

WORKDIR /data/bcs/bcs-api/
CMD ["/data/bcs/bcs-api/container-start.sh"]


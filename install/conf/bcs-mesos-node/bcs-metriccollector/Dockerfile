FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-metriccollector

ADD bcs-metriccollector /data/bcs/bcs-metriccollector/
ADD bcs-metriccollector.json.template /data/bcs/bcs-metriccollector/
ADD container-start.sh /data/bcs/bcs-metriccollector/
RUN chmod +x /data/bcs/bcs-metriccollector/container-start.sh

WORKDIR /data/bcs/bcs-metriccollector/
CMD ["/data/bcs/bcs-metriccollector/container-start.sh"]


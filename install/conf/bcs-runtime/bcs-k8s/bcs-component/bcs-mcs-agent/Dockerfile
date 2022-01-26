FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-mcs-agent/

ADD bcs-mcs-agent /data/bcs/bcs-mcs-agent/

WORKDIR /data/bcs/bcs-mcs-agent/
CMD [ "/data/bcs/bcs-mcs-agent/bcs-mcs-agent" ]

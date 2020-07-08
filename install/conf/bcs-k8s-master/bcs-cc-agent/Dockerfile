FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cc-agent/

ADD bcs-cc-agent /data/bcs/bcs-cc-agent/
ADD container-start.sh /data/bcs/bcs-cc-agent/

RUN chmod +x /data/bcs/bcs-cc-agent/container-start.sh

WORKDIR /data/bcs/bcs-cc-agent/
CMD [ "/data/bcs/bcs-cc-agent/container-start.sh" ]

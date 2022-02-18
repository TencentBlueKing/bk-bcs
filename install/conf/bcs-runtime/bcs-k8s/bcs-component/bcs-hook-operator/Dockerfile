FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-hook-operator/

ADD bcs-hook-operator /data/bcs/bcs-hook-operator/
ADD container-start.sh /data/bcs/bcs-hook-operator/

RUN chmod +x /data/bcs/bcs-hook-operator/bcs-hook-operator
RUN chmod +x /data/bcs/bcs-hook-operator/container-start.sh

WORKDIR /data/bcs/bcs-hook-operator/
CMD [ "/data/bcs/bcs-hook-operator/container-start.sh" ]

FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-hook-daemonset/

ADD bcs-hook-daemonset /data/bcs/bcs-hook-daemonset/
ADD container-start.sh /data/bcs/bcs-hook-daemonset/

RUN chmod +x /data/bcs/bcs-hook-daemonset/container-start.sh

WORKDIR /data/bcs/bcs-hook-daemonset/
CMD [ "/data/bcs/bcs-hook-daemonset/container-start.sh" ]

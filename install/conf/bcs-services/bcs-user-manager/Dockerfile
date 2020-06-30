FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-user-manager

ADD bcs-user-manager /data/bcs/bcs-user-manager/
ADD container-start.sh /data/bcs/bcs-user-manager/
RUN chmod +x /data/bcs/bcs-user-manager/container-start.sh

WORKDIR /data/bcs/bcs-user-manager/
ENTRYPOINT ["/data/bcs/bcs-user-manager/container-start.sh"]


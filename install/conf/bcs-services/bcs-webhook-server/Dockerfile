FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-webhook-server

ADD bcs-webhook-server /data/bcs/bcs-webhook-server/
ADD container-start.sh /data/bcs/bcs-webhook-server/
RUN chmod +x /data/bcs/bcs-webhook-server/container-start.sh

WORKDIR /data/bcs/bcs-webhook-server/
ENTRYPOINT ["/data/bcs/bcs-webhook-server/container-start.sh"]


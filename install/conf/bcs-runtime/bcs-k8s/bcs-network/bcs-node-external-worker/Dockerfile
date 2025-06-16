FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-node-external-worker

ADD bcs-node-external-worker /data/bcs/bcs-node-external-worker/
ADD container-start.sh /data/bcs/bcs-node-external-worker/
RUN chmod +x /data/bcs/bcs-node-external-worker/container-start.sh

WORKDIR /data/bcs/bcs-node-external-worker/
CMD ["/data/bcs/bcs-node-external-worker/container-start.sh"]

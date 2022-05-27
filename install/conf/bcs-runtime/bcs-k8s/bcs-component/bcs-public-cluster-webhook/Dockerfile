FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-public-cluster-webhook/

ADD bcs-public-cluster-webhook /data/bcs/bcs-public-cluster-webhook/
ADD container-start.sh /data/bcs/bcs-public-cluster-webhook/

RUN chmod +x /data/bcs/bcs-public-cluster-webhook/container-start.sh
RUN chmod +x /data/bcs/bcs-public-cluster-webhook/bcs-public-cluster-webhook
WORKDIR /data/bcs/bcs-public-cluster-webhook/
CMD [ "/data/bcs/bcs-public-cluster-webhook/container-start.sh" ]

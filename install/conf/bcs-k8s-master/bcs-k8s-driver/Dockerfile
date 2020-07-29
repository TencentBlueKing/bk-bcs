FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-driver

ADD bcs-k8s-driver /data/bcs/bcs-k8s-driver/
RUN chmod +x /data/bcs/bcs-k8s-driver/bcs-k8s-driver

WORKDIR /data/bcs/bcs-k8s-driver/
ENTRYPOINT ["/data/bcs/bcs-k8s-driver/bcs-k8s-driver"]
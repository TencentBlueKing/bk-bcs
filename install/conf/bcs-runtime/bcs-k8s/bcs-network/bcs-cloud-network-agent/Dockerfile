FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/cert && mkdir -p /data/bcs/logs/bcs && mkdir -p /bcs/cni && mkdir -p /bcs/etc
RUN mkdir -p /data/bcs/bcs-cloud-network-agent /data/bcs/bcs-cni
COPY ./bcs-cloud-network-agent /data/bcs/bcs-cloud-network-agent/
COPY ./bcs-cloud-network-agent.conf.template /data/bcs/bcs-cloud-network-agent/
COPY ./bcs-eni /data/bcs/bcs-cni/
COPY ./bcs-eni.conf /data/bcs/bcs-cni/bcs-eni.conf
COPY ./container-start.sh /data/bcs/bcs-cloud-network-agent/container-start.sh
RUN chmod +x /data/bcs/bcs-cloud-network-agent/container-start.sh && chmod +x /data/bcs/bcs-cni/bcs-eni 
WORKDIR /data/bcs

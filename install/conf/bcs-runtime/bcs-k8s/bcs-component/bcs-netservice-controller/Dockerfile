FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

WORKDIR /root

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/kubeconfigs /data/bcs/cni/bin /data/bcs/cni/conf

COPY ./bcs-underlay-cni ./bcs-netservice-ipam /data/bcs/cni/bin/
ADD bcs-netservice-controller /data/bcs/bcs-netservice-controller/

WORKDIR /data/bcs/bcs-netservice-controller/

CMD ["/data/bcs/bcs-netservice-controller/bcs-netservice-controller"]

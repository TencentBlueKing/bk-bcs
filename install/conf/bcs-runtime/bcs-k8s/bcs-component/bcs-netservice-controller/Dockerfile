FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

ENV TZ Asia/Shanghai
WORKDIR /root
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/kubeconfigs /data/bcs/cni/bin /data/bcs/cni/conf

COPY ./bcs-underlay-cni ./bcs-netservice-ipam /data/bcs/cni/bin/
ADD bcs-netservice-controller /data/bcs/bcs-netservice-controller/

WORKDIR /data/bcs/bcs-netservice-controller/

CMD ["/data/bcs/bcs-netservice-controller/bcs-netservice-controller"]

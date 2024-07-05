FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-csi-tencentcloud/

COPY  bcs-k8s-csi-tencentcloud /data/bcs/bcs-k8s-csi-tencentcloud/

WORKDIR /data/bcs/bcs-k8s-csi-tencentcloud/
CMD ["/data/bcs/bcs-k8s-csi-tencentcloud/bcs-k8s-csi-tencentcloud"]

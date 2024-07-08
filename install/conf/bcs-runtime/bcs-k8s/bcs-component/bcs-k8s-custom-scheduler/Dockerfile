FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-custom-scheduler 
ADD ./bcs-k8s-custom-scheduler /data/bcs/bcs-k8s-custom-scheduler/

ENTRYPOINT ["/data/bcs/bcs-k8s-custom-scheduler/bcs-k8s-custom-scheduler"]

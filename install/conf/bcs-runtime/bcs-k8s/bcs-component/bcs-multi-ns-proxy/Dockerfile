FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/kubeconfigs

ADD bcs-multi-ns-proxy /data/bcs/bcs-multi-ns-proxy/

WORKDIR /data/bcs/bcs-multi-ns-proxy/

CMD ["/data/bcs/bcs-multi-ns-proxy/bcs-multi-ns-proxy"]

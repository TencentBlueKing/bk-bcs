FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-kube-agent

ADD bcs-kube-agent /data/bcs/bcs-kube-agent/
RUN chmod +x /data/bcs/bcs-kube-agent/bcs-kube-agent

WORKDIR /data/bcs/bcs-kube-agent/
ENTRYPOINT ["/data/bcs/bcs-kube-agent/bcs-kube-agent"]

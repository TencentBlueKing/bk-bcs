FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-argocd-manager/bcs-argocd-proxy

ADD bcs-argocd-proxy /data/bcs/bcs-argocd-manager/bcs-argocd-proxy/
ADD container-start.sh /data/bcs/bcs-argocd-manager/bcs-argocd-proxy/
RUN chmod +x /data/bcs/bcs-argocd-manager/bcs-argocd-proxy/container-start.sh

WORKDIR /data/bcs/bcs-argocd-manager/bcs-argocd-proxy/
CMD ["/data/bcs/bcs-argocd-manager/bcs-argocd-proxy/container-start.sh"]

FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs
RUN mkdir -p /data/bcs/bcs-argocd-manager/bcs-argocd-controller

ADD bcs-argocd-controller /data/bcs/bcs-argocd-manager/bcs-argocd-controller/
ADD container-start.sh /data/bcs/bcs-argocd-manager/bcs-argocd-controller/
ADD charts /data/bcs/bcs-argocd-manager/bcs-argocd-controller/charts/
RUN chmod +x /data/bcs/bcs-argocd-manager/bcs-argocd-controller/container-start.sh

WORKDIR /data/bcs/bcs-argocd-manager/bcs-argocd-controller/
CMD ["/data/bcs/bcs-argocd-manager/bcs-argocd-controller/container-start.sh"]

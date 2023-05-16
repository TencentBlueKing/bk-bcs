FROM centos:7

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-argocd-manager/bcs-argocd-server

ADD bcs-argocd-server /data/bcs/bcs-argocd-manager/bcs-argocd-server/
ADD container-start.sh /data/bcs/bcs-argocd-manager/bcs-argocd-server/
RUN chmod +x /data/bcs/bcs-argocd-manager/bcs-argocd-server/container-start.sh

WORKDIR /data/bcs/bcs-argocd-manager/bcs-argocd-server/
CMD ["/data/bcs/bcs-argocd-manager/bcs-argocd-server/container-start.sh"]

FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

RUN mkdir -p /data/bcs/bcs-project-manager-migration

ADD bcs-project-manager-migration /data/bcs/bcs-project-manager-migration/
RUN chmod +x /data/bcs/bcs-project-manager-migration/bcs-project-manager-migration

WORKDIR /data/bcs/bcs-project-manager-migration/
CMD ["/data/bcs/bcs-project-manager-migration/bcs-project-manager-migration"]

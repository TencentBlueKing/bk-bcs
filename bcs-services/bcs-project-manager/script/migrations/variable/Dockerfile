FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

RUN mkdir -p /data/bcs/bcs-variable-migration

ADD bcs-variable-migration /data/bcs/bcs-variable-migration/
RUN chmod +x /data/bcs/bcs-variable-migration/bcs-variable-migration

WORKDIR /data/bcs/bcs-variable-migration/
CMD ["/data/bcs/bcs-variable-migration/bcs-variable-migration"]

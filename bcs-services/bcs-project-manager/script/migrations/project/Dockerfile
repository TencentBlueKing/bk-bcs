FROM centos:7

RUN mkdir -p /data/bcs/bcs-project-manager-migration

ADD bcs-project-manager-migration /data/bcs/bcs-project-manager-migration/
RUN chmod +x /data/bcs/bcs-project-manager-migration/bcs-project-manager-migration

WORKDIR /data/bcs/bcs-project-manager-migration/
CMD ["/data/bcs/bcs-project-manager-migration/bcs-project-manager-migration"]

FROM centos:7

RUN mkdir -p /data/bcs/bcs-variable-migration

ADD bcs-variable-migration /data/bcs/bcs-variable-migration/
RUN chmod +x /data/bcs/bcs-variable-migration/bcs-variable-migration

WORKDIR /data/bcs/bcs-variable-migration/
CMD ["/data/bcs/bcs-variable-migration/bcs-variable-migration"]

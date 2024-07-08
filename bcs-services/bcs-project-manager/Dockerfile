FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

LABEL maintainer="Tencent BlueKing"

RUN mkdir -p /data/bcs/logs /data/bcs/cert /data/bcs/swagger
ADD third_party/swagger-ui/* /data/bcs/swagger/
ADD proto/bcsproject/bcsproject.swagger.json /data/bcs/swagger/bcsproject.swagger.json

WORKDIR /data/bcs/

ADD bcs-project-manager /usr/bin/bcs-project-manager
ADD bcs-project-migration /usr/bin/bcs-project-migration
ADD bcs-variable-migration /usr/bin/bcs-variable-migration

RUN chmod +x /usr/bin/bcs-project-manager

ENV GODEBUG x509sha1=1

ENTRYPOINT ["/usr/bin/bcs-project-manager"]
CMD ["-c", "/data/bcs/bcs-project-manager/bcs-project-manager.yaml"]

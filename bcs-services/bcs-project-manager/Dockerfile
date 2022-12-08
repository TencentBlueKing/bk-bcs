FROM centos:7

LABEL maintainer="Tencent BlueKing"

RUN mkdir -p /data/bcs/logs /data/bcs/cert /data/bcs/swagger
ADD third_party/swagger-ui/* /data/bcs/swagger/
ADD proto/bcsproject/bcsproject.swagger.json /data/bcs/swagger/bcsproject.swagger.json

WORKDIR /data/bcs/

ADD bcs-project-manager /usr/bin/bcs-project-manager
ADD bcs-project-migration /usr/bin/bcs-project-migration

RUN chmod +x /usr/bin/bcs-project-manager

ENTRYPOINT ["/usr/bin/bcs-project-manager"]
CMD ["-c", "/data/bcs/bcs-project-manager/bcs-project-manager.yaml"]
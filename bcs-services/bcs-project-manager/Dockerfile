FROM centos:7

LABEL maintainer="Tencent BlueKing"

RUN mkdir -p /data/bcs/logs /data/bcs/cert /data/bcs/swagger
ADD swagger/swagger-ui/* /data/bcs/swagger/
ADD swagger/bcsproject.swagger.json /data/bcs/swagger/bcsproject.swagger.json

WORKDIR /data/bcs/

ADD bcs-project-manager /usr/bin/bcs-project-manager

RUN chmod +x /usr/bin/bcs-project-manager

ENTRYPOINT ["/usr/bin/bcs-project-manager"]
CMD ["-c", "/data/bcs/bcs-project-manager/bcs-project-manager.yaml"]
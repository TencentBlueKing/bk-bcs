FROM tencentos/tencentos4-minimal

LABEL maintainer="Tencent BlueKing"

RUN mkdir -p /data/bcs/logs /data/bcs/cert /data/bcs/swagger
ADD swagger/swagger-ui/* /data/bcs/swagger/
ADD swagger/bcsproject.swagger.json /data/bcs/swagger/bcsproject.swagger.json

WORKDIR /data/bcs/

ADD bcs-project-migration /usr/bin/bcs-project-migration
ADD bcs-project-manager /usr/bin/bcs-project-manager
ADD bcs-variable-migration /usr/bin/bcs-variable-migration

RUN chmod +x /usr/bin/bcs-project-manager

ENV GODEBUG x509sha1=1

ENTRYPOINT ["/usr/bin/bcs-project-manager"]
CMD ["-c", "/data/bcs/bcs-project-manager/bcs-project-manager.yaml"]

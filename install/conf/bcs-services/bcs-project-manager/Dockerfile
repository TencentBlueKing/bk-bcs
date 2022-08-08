FROM centos:7

LABEL maintainer="Tencent BlueKing"

RUN mkdir -p /data/bcs/logs /data/bcs/cert /data/bcs/swagger

ADD bcs-project-manager /usr/bin/bcs-project-manager

RUN chmod +x /usr/bin/bcs-project-manager

ENTRYPOINT ["/bin/bash", "-c", "/usr/bin/bcs-project-manager"]
CMD ["-c", "/data/bcs/bcs-project-manager/bcs-project-manager.yaml"]
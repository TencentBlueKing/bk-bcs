FROM tencentos/tencentos4-minimal

ENV GODEBUG=x509sha1=1

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-helm-manager
RUN mkdir -p /data/bcs/bcs-helm-manager/runtime

ADD bcs-helm-manager /data/bcs/bcs-helm-manager/
ADD bcs-helm-manager-migrator /data/bcs/bcs-helm-manager/
ADD container-start.sh /data/bcs/bcs-helm-manager/
ADD swagger/swagger-ui /data/bcs/swagger
ADD lc_msgs.yaml /data/bcs/bcs-helm-manager/lc_msgs.yaml
RUN chmod +x /data/bcs/bcs-helm-manager/container-start.sh

WORKDIR /data/bcs/bcs-helm-manager/
CMD ["/data/bcs/bcs-helm-manager/container-start.sh"]

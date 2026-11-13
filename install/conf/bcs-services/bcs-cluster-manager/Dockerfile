FROM tencentos/tencentos4-minimal

RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-cluster-manager

ADD swagger/swagger-ui /data/bcs/swagger
ADD bcs-cluster-manager /data/bcs/bcs-cluster-manager/
ADD bcs-cluster-manager.json.template /data/bcs/bcs-cluster-manager/
ADD container-start.sh /data/bcs/bcs-cluster-manager/
RUN chmod +x /data/bcs/bcs-cluster-manager/container-start.sh

WORKDIR /data/bcs/bcs-cluster-manager/
CMD ["/data/bcs/bcs-cluster-manager/container-start.sh"]

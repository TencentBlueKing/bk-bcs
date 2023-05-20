FROM centos:7

ENV GODEBUG=x509sha1=1

RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-helm-manager
RUN mkdir -p /data/bcs/bcs-helm-manager/runtime

ADD bcs-helm-manager /data/bcs/bcs-helm-manager/
ADD bcs-helm-manager-migrator /data/bcs/bcs-helm-manager/
ADD container-start.sh /data/bcs/bcs-helm-manager/
ADD swagger/swagger-ui /data/bcs/swagger
RUN chmod +x /data/bcs/bcs-helm-manager/container-start.sh

WORKDIR /data/bcs/bcs-helm-manager/
CMD ["/data/bcs/bcs-helm-manager/container-start.sh"]

FROM centos:8

# for command envsubset
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/swagger
RUN mkdir -p /data/bcs/bcs-cluster-manager

ADD swagger/swagger-ui /data/bcs/swagger
ADD bcs-cluster-manager /data/bcs/bcs-cluster-manager/
ADD bcs-cluster-manager.json.template /data/bcs/bcs-cluster-manager/
ADD container-start.sh /data/bcs/bcs-cluster-manager/
RUN chmod +x /data/bcs/bcs-cluster-manager/container-start.sh

ENV TZ="Asia/Shanghai"
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

WORKDIR /data/bcs/bcs-cluster-manager/
CMD ["/data/bcs/bcs-cluster-manager/container-start.sh"]

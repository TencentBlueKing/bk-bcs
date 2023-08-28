FROM centos:8

#for command envsubst
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-user-manager

ADD bcs-user-manager /data/bcs/bcs-user-manager/
ADD bcs-user-manager.json.template /data/bcs/bcs-user-manager/
ADD container-start.sh /data/bcs/bcs-user-manager/
RUN chmod +x /data/bcs/bcs-user-manager/container-start.sh

ENV TZ="Asia/Shanghai"
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

WORKDIR /data/bcs/bcs-user-manager/
CMD ["/data/bcs/bcs-user-manager/container-start.sh"]


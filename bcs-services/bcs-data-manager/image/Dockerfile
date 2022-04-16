FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/bcs-data-manager/
RUN mkdir -p /data/bcs/logs/bcs
RUN mkdir -p /data/bcs/bcs-cluster-manager

ADD bcs-data-manager /data/bcs/bcs-data-manager/
ADD bcs-data-manager.json.template /data/bcs/bcs-data-manager/
ADD container-start.sh /data/bcs/bcs-data-manager/

RUN chmod +x /data/bcs/bcs-data-manager/container-start.sh

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf

ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

WORKDIR /data/bcs/bcs-data-manager/
CMD [ "/data/bcs/bcs-data-manager/container-start.sh" ]
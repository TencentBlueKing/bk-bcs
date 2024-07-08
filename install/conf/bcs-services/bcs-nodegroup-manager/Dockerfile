FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-nodegroup-manager/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-nodegroup-manager /data/bcs/bcs-nodegroup-manager/
ADD bcs-nodegroup-manager.json.template /data/bcs/bcs-nodegroup-manager/
ADD container-start.sh /data/bcs/bcs-nodegroup-manager/

RUN chmod +x /data/bcs/bcs-nodegroup-manager/container-start.sh
RUN chmod +x /data/bcs/bcs-nodegroup-manager/bcs-nodegroup-manager


WORKDIR /data/bcs/bcs-nodegroup-manager/
CMD [ "/data/bcs/bcs-nodegroup-manager/container-start.sh" ]

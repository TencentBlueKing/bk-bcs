FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-vaultplugin-server /data/bcs/bcs-gitops-vaultplugin-server/
ADD bcs-gitops-vaultplugin-server.json.template /data/bcs/bcs-gitops-vaultplugin-server/
ADD container-start.sh /data/bcs/bcs-gitops-vaultplugin-server/

RUN chmod +x /data/bcs/bcs-gitops-vaultplugin-server/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-vaultplugin-server/bcs-gitops-vaultplugin-server

WORKDIR /data/bcs/bcs-gitops-vaultplugin-server/
CMD [ "/data/bcs/bcs-gitops-vaultplugin-server/container-start.sh" ]

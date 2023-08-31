FROM centos:7


RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-vaultplugin-server /data/bcs/bcs-gitops-vaultplugin-server/
ADD bcs-gitops-vaultplugin-server.json /data/bcs/bcs-gitops-vaultplugin-server/
ADD container-start.sh /data/bcs/bcs-gitops-vaultplugin-server/

RUN chmod +x /data/bcs/bcs-gitops-vaultplugin-server/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-vaultplugin-server/bcs-gitops-vaultplugin-server


WORKDIR /data/bcs/bcs-gitops-vaultplugin-server/
CMD [ "/data/bcs/bcs-gitops-vaultplugin-server/container-start.sh" ]
FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-gitops-gitgenerator-webhook/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-gitgenerator-webhook /data/bcs/bcs-gitops-gitgenerator-webhook/
ADD bcs-gitops-gitgenerator-webhook.json.template /data/bcs/bcs-gitops-gitgenerator-webhook/
ADD container-start.sh /data/bcs/bcs-gitops-gitgenerator-webhook/

RUN chmod +x /data/bcs/bcs-gitops-gitgenerator-webhook/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-gitgenerator-webhook/bcs-gitops-gitgenerator-webhook

WORKDIR /data/bcs/bcs-gitops-gitgenerator-webhook/
CMD ["/data/bcs/bcs-gitops-gitgenerator-webhook/container-start.sh"]

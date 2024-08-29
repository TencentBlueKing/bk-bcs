FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-gitops-pre-check/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-pre-check /data/bcs/bcs-gitops-pre-check/
ADD bcs-gitops-pre-check.json.template /data/bcs/bcs-gitops-pre-check/
ADD container-start.sh /data/bcs/bcs-gitops-pre-check/

RUN chmod +x /data/bcs/bcs-gitops-pre-check/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-pre-check/bcs-gitops-pre-check


WORKDIR /data/bcs/bcs-gitops-pre-check/
CMD [ "/data/bcs/bcs-gitops-pre-check/container-start.sh" ]
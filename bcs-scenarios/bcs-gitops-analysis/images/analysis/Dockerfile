FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext
## yum install epel-release -y && yum install dsniff -y && yum install net-tools -y

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-analysis /data/bcs/bcs-gitops-analysis/
ADD bcs-gitops-analysis.json.template /data/bcs/bcs-gitops-analysis/
ADD container-start.sh /data/bcs/bcs-gitops-analysis/

RUN chmod +x /data/bcs/bcs-gitops-analysis/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-analysis/bcs-gitops-analysis

WORKDIR /data/bcs/bcs-gitops-analysis/
CMD [ "/data/bcs/bcs-gitops-analysis/container-start.sh" ]

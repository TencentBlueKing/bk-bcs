FROM centos:7
RUN sed -i s/mirror.centos.org/mirrors.tencent.com/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-gitops-proxy/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-proxy /data/bcs/bcs-gitops-proxy/
ADD bcs-gitops-proxy.json.template /data/bcs/bcs-gitops-proxy/
ADD container-start.sh /data/bcs/bcs-gitops-proxy/

RUN chmod +x /data/bcs/bcs-gitops-proxy/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-proxy/bcs-gitops-proxy


WORKDIR /data/bcs/bcs-gitops-proxy/
CMD [ "/data/bcs/bcs-gitops-proxy/container-start.sh" ]

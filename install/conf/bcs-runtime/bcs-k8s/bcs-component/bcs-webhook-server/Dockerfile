FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/plugins

ADD bcs-webhook-server /data/bcs/bcs-webhook-server/
ADD container-start.sh /data/bcs/bcs-webhook-server/
RUN chmod +x /data/bcs/bcs-webhook-server/container-start.sh

WORKDIR /data/bcs/bcs-webhook-server/
CMD ["/data/bcs/bcs-webhook-server/container-start.sh"]


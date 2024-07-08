FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-hook-operator/

ADD bcs-hook-operator /data/bcs/bcs-hook-operator/
ADD container-start.sh /data/bcs/bcs-hook-operator/

RUN chmod +x /data/bcs/bcs-hook-operator/bcs-hook-operator
RUN chmod +x /data/bcs/bcs-hook-operator/container-start.sh

WORKDIR /data/bcs/bcs-hook-operator/
CMD [ "/data/bcs/bcs-hook-operator/container-start.sh" ]

FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-gamestatefulset-operator/

ADD bcs-gamestatefulset-operator /data/bcs/bcs-gamestatefulset-operator/
ADD container-start.sh /data/bcs/bcs-gamestatefulset-operator/

RUN chmod +x /data/bcs/bcs-gamestatefulset-operator/bcs-gamestatefulset-operator
RUN chmod +x /data/bcs/bcs-gamestatefulset-operator/container-start.sh

WORKDIR /data/bcs/bcs-gamestatefulset-operator/
CMD [ "/data/bcs/bcs-gamestatefulset-operator/container-start.sh" ]

FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-bkcmdb-synchronizer

ADD bcs-bkcmdb-synchronizer /data/bcs/bcs-bkcmdb-synchronizer/
ADD bcs-bkcmdb-synchronizer.json.template /data/bcs/bcs-bkcmdb-synchronizer/
ADD container-start.sh /data/bcs/bcs-bkcmdb-synchronizer/

RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer
RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/container-start.sh

WORKDIR /data/bcs/bcs-bkcmdb-synchronizer/

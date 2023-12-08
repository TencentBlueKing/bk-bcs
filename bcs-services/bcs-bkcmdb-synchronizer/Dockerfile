FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-bkcmdb-synchronizer

ADD bin/bcs-bkcmdb-synchronizer /data/bcs/bcs-bkcmdb-synchronizer/
ADD config/bcs-bkcmdb-synchronizer.json.template /data/bcs/bcs-bkcmdb-synchronizer/
ADD config/container-start.sh /data/bcs/bcs-bkcmdb-synchronizer/

RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer
RUN chmod +x /data/bcs/bcs-bkcmdb-synchronizer/container-start.sh

WORKDIR /data/bcs/bcs-bkcmdb-synchronizer/
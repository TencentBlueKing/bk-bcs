FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-storage

ADD bcs-storage /data/bcs/bcs-storage/
ADD container-start.sh /data/bcs/bcs-storage/
ADD bcs-storage.json.template /data/bcs/bcs-storage/
ADD storage-database.conf.template /data/bcs/bcs-storage/
ADD queue.conf.template /data/bcs/bcs-storage/
RUN chmod +x /data/bcs/bcs-storage/container-start.sh

WORKDIR /data/bcs/bcs-storage/
CMD ["/data/bcs/bcs-storage/container-start.sh"]


FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-dns-service

ADD bcs-dns-service /data/bcs/bcs-dns-service/
ADD bcs-dns-service.config.template /data/bcs/bcs-dns-service/
ADD container-start.sh /data/bcs/bcs-dns-service/
RUN chmod +x /data/bcs/bcs-dns-service/container-start.sh

WORKDIR /data/bcs/bcs-dns-service/
CMD ["/data/bcs/bcs-dns-service/container-start.sh"]


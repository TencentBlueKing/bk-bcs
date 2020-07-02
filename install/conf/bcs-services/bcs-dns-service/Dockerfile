FROM centos:latest

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-dns-service

ADD bcs-dns-service /data/bcs/bcs-dns-service/
ADD container-start.sh /data/bcs/bcs-dns-service/
RUN chmod +x /data/bcs/bcs-dns-service/container-start.sh

WORKDIR /data/bcs/bcs-dns-service/
ENTRYPOINT ["/data/bcs/bcs-dns-service/container-start.sh"]


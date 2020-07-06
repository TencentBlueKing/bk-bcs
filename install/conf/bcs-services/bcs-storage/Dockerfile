FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-storage

ADD bcs-storage /data/bcs/bcs-storage/
ADD container-start.sh /data/bcs/bcs-storage/
RUN chmod +x /data/bcs/bcs-storage/container-start.sh

WORKDIR /data/bcs/bcs-storage/
ENTRYPOINT ["/data/bcs/bcs-storage/container-start.sh"]


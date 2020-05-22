FROM centos:latest

RUN mkdir -p /data/bcs/bcs-storage && mkdir -p /data/bcs/logs/bcs && mkdir -p /data/bcs/cert
COPY bcs-storage /data/bcs/bcs-storage
COPY bcs-storage.json.template /data/bcs/bcs-storage/bcs-storage.json
RUN chmod +x /data/bcs/bcs-storage/bcs-storage
WORKDIR /data/bcs/bcs-storage
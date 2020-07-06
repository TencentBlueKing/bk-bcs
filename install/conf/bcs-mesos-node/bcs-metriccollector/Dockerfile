FROM centos:7

RUN mkdir -p /data/bcs/logs

COPY start_docker.sh /data/bcs/start_docker.sh
COPY config_file_docker.json /data/bcs/config_file_docker.json
COPY bcs-metriccollector /data/bcs/bcs-services/bcs-metriccollector
COPY cert /data/bcs/cert

ENTRYPOINT /data/bcs/start_docker.sh
FROM centos:latest

COPY ./bcs-k8s-watch /
COPY ./cert /cert

ENTRYPOINT ["/bcs-k8s-watch", "--config", "/config/config.json"]

FROM centos:latest

COPY ./bcs-k8s-watch /
RUN mkdir /cert
COPY bcs-* /cert

ENTRYPOINT ["/bcs-k8s-watch", "--config", "/config/config.json"]

FROM dockerhub.com:8443/public/k8s/base/debian:jessie

COPY ./k8s-watch /
COPY ./cert /cert

ENTRYPOINT ["/k8s-watch", "--config", "/config/config.json"]

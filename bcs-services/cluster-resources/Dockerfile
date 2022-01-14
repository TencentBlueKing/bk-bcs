FROM centos:7

LABEL maintainer="Tencent BlueKing"

WORKDIR /data/workspace

ADD . .

RUN cp ./cluster-resources-service /usr/local/bin && mkdir -p /tmp/logs

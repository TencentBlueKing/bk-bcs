# ---------------- runner container ----------------

FROM alpine:3.15

LABEL maintainer="Tencent BlueKing"

RUN apk --update --no-cache add bash ca-certificates vim

WORKDIR /data/workspace

COPY cluster-resources-service /usr/bin/cluster-resources-service

COPY etc/ /data/workspace/etc

COPY swagger/ /data/workspace/swagger

# package example files 

RUN mkdir -p /data/workspace/example

COPY example/config/ /data/workspace/example/config

COPY example/manifest/ /data/workspace/example/manifest

COPY example/reference/ /data/workspace/example/reference

ENV EXAMPLE_FILE_BASE_DIR=/data/workspace/example

# package form tmpl & schema files

RUN mkdir -p /data/workspace/form_tmpl

COPY tmpl/ /data/workspace/form_tmpl

ENV FORM_TMPL_FILE_BASE_DIR=/data/workspace/form_tmpl

# package i18n file

COPY lc_msgs.yaml /data/workspace/lc_msgs.yaml

ENV LOCALIZE_FILE_PATH=/data/workspace/lc_msgs.yaml

# ---------------- runner container ----------------

FROM tencentos/tencentos4-minimal

LABEL maintainer="Tencent BlueKing"

WORKDIR /data/workspace

COPY bcs-cluster-resources /usr/bin/bcs-cluster-resources

COPY etc/ /data/workspace/etc

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

ENV GODEBUG x509sha1=1

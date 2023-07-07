# -------------- builder container --------------
FROM golang:1.17.5 AS builder

LABEL maintainer="Tencent BlueKing"

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /go/src/

ARG VERSION
ARG GITCOMMIT

RUN apt-get install make

COPY . .
RUN go mod download
RUN make build VERSION=$VERSION GITCOMMIT=$GITCOMMIT

# package example files
ARG SRC_EXAMPLE_DIR=/go/src/pkg/resource/example
ARG TARGET_EXAMPLE_DIR=/go/src/example

RUN mkdir -p $TARGET_EXAMPLE_DIR/config
RUN mkdir -p $TARGET_EXAMPLE_DIR/manifest
RUN mkdir -p $TARGET_EXAMPLE_DIR/reference

RUN cp -R $SRC_EXAMPLE_DIR/config/* $TARGET_EXAMPLE_DIR/config/
RUN cp -R $SRC_EXAMPLE_DIR/manifest/* $TARGET_EXAMPLE_DIR/manifest/
RUN cp -R $SRC_EXAMPLE_DIR/reference/* $TARGET_EXAMPLE_DIR/reference/

# package form tmpl & schema files
ARG SRC_FORM_TMPL_DIR=/go/src/pkg/resource/form/tmpl
ARG TARGET_FORM_TMPL_DIR=/go/src/form_tmpl

RUN mkdir -p $TARGET_FORM_TMPL_DIR

RUN cp -R $SRC_FORM_TMPL_DIR/* $TARGET_FORM_TMPL_DIR/

# -------------- runner container --------------
FROM alpine:3.15 AS runner

LABEL maintainer="Tencent BlueKing"

RUN apk --update --no-cache add bash ca-certificates vim

WORKDIR /data/workspace

COPY --from=builder /go/src/bcs-cluster-resources /usr/bin/bcs-cluster-resources

COPY --from=builder /go/src/etc /data/workspace/etc

COPY --from=builder /go/src/example /data/workspace/example

ENV EXAMPLE_FILE_BASE_DIR=/data/workspace/example

COPY --from=builder /go/src/form_tmpl /data/workspace/form_tmpl

ENV FORM_TMPL_FILE_BASE_DIR=/data/workspace/form_tmpl

COPY --from=builder /go/src/pkg/i18n/locale/lc_msgs.yaml /data/workspace/lc_msgs.yaml

ENV LOCALIZE_FILE_PATH=/data/workspace/lc_msgs.yaml

ENV GODEBUG x509sha1=1

# default log file base dir
RUN mkdir -p /tmp/logs

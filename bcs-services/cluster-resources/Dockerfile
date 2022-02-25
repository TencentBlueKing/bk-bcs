# -------------- builder container --------------
FROM golang:1.14.15 AS builder

LABEL maintainer="Tencent BlueKing"

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /go/src/

ARG VERSION
ARG GITCOMMIT

RUN apt-get install make

COPY . .
RUN go mod download
RUN make build VERSION=$VERSION GITCOMMIT=$GITCOMMIT

# package swagger dependency file
RUN cp -R ./third_party/swagger-ui/* ./swagger/

# package example file
ARG SRC_EXAMPLE_DIR=/go/src/pkg/resource/example
ARG TARGET_EXAMPLE_DIR=/go/src/example

RUN mkdir -p $TARGET_EXAMPLE_DIR/config
RUN mkdir -p $TARGET_EXAMPLE_DIR/manifest
RUN mkdir -p $TARGET_EXAMPLE_DIR/reference

RUN cp -R $SRC_EXAMPLE_DIR/config/* $TARGET_EXAMPLE_DIR/config/
RUN cp -R $SRC_EXAMPLE_DIR/manifest/* $TARGET_EXAMPLE_DIR/manifest/
RUN cp -R $SRC_EXAMPLE_DIR/reference/* $TARGET_EXAMPLE_DIR/reference/

# -------------- runner container --------------
FROM alpine:3.15 AS runner

LABEL maintainer="Tencent BlueKing"

RUN apk --update --no-cache add bash ca-certificates

WORKDIR /data/workspace

COPY --from=builder /go/src/cluster-resources-service /usr/bin/cluster-resources-service

COPY --from=builder /go/src/etc /data/workspace/etc

COPY --from=builder /go/src/swagger /data/workspace/swagger

COPY --from=builder /go/src/example /data/workspace/example

ENV EXAMPLE_FILE_BASE_DIR=/data/workspace/example

# default log file base dir
RUN mkdir -p /tmp/logs

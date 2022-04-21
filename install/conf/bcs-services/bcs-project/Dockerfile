FROM alpine:3.15

LABEL maintainer="Tencent BlueKing"

RUN apk --update --no-cache add bash ca-certificates

WORKDIR /data/project

RUN mkdir -p /data/project/logs /data/project/cert /data/project/swagger

COPY bcs-project-service /usr/bin/bcs-project-service
COPY ./swagger /data/project/swagger

CMD ["/usr/bin/bcs-project-service"]

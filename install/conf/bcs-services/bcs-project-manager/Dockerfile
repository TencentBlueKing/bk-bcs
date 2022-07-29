FROM alpine:3.15

LABEL maintainer="Tencent BlueKing"

RUN apk --update --no-cache add bash ca-certificates vim curl

WORKDIR /data/bcs

RUN mkdir -p /data/bcs/logs /data/bcs/cert /data/bcs/swagger

COPY bcs-project-manager /usr/bin/bcs-project-manager

CMD ["/usr/bin/bcs-project-manager"]

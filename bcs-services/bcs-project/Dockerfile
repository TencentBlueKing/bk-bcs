# -------------- builder container --------------
FROM golang:1.17.5 AS builder

LABEL maintainer="Tencent BlueKing"

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /go/src/

ARG VERSION

RUN apt-get install make

COPY . .
RUN go mod download
RUN make build VERSION=$VERSION

# swagger
RUN mkdir -p ./swagger && cp -R ./third_party/swagger-ui/* ./swagger/ \
    && cp ./proto/bcsproject/bcsproject.swagger.json ./swagger/

# -------------- runner container --------------
FROM alpine:3.15 AS runner

LABEL maintainer="Tencent BlueKing"

RUN apk --update --no-cache add bash ca-certificates

WORKDIR /data/project

COPY --from=builder /go/src/bcs-project-service /usr/bin/bcs-project-service
COPY --from=builder /go/src/etc /data/project/etc
COPY --from=builder /go/src/swagger /data/project/

RUN mkdir -p /data/project/logs
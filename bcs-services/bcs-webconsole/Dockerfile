FROM golang:alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/bcs-webconsole
RUN apk --update --no-cache add ca-certificates gcc libtool make musl-dev protoc
COPY Makefile go.mod go.sum ./
RUN make init && go mod download
COPY . .
RUN make proto tidy build

FROM scratch
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /go/src/bcs-webconsole/bcs-webconsole /bcs-webconsole
ENTRYPOINT ["/bcs-webconsole"]
CMD []

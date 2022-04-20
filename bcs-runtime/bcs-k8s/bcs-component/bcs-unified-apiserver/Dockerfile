FROM golang:alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/bcs-unified-apiserver
RUN apk --update --no-cache add ca-certificates gcc libtool make musl-dev protoc
COPY Makefile go.mod go.sum ./
RUN go mod download
COPY . .
RUN make tidy build

FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim
COPY --from=builder /go/src/bcs-unified-apiserver/_output/bcs-unified-apiserver /bcs-unified-apiserver
ENTRYPOINT ["/bcs-unified-apiserver"]
CMD []

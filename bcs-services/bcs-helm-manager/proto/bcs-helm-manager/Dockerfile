FROM golang:1.18 AS builder

ARG PROTOC_GEN_GO_VERSION=1.3.2
ARG PROTOC_GEN_MICRO_VERSION=1.0.0
ARG PROTOC_GEN_GRPC_GATEWAY_VERSION=1.14.6
ARG PROTOC_GEN_SWAGGER_VERSION=1.14.6
ARG PROTOC_GEN_VALIDATE_VERSION=0.6.7

# Install protoc-gen-go
RUN go install github.com/golang/protobuf/protoc-gen-go@v${PROTOC_GEN_GO_VERSION}

# Install protoc-gen-micro
RUN go install github.com/go-micro/generator/cmd/protoc-gen-micro@v${PROTOC_GEN_MICRO_VERSION}

# Install protoc-gen-gateway
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${PROTOC_GEN_GRPC_GATEWAY_VERSION}

# Install protoc-gen-swagger
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${PROTOC_GEN_SWAGGER_VERSION}

# Install protoc-gen-validate
RUN go install github.com/envoyproxy/protoc-gen-validate@v${PROTOC_GEN_VALIDATE_VERSION}

FROM debian:bullseye-slim

WORKDIR /proto

ARG PROTOC_VERSION=3.12.3

RUN apt update && apt install wget unzip -y \
    && rm -rf /var/lib/apt/lists/*

RUN wget -q https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip -O protoc.zip \
    && unzip protoc.zip -d /usr/local \
    && rm protoc.zip

COPY --from=builder /go/bin/ /usr/local/bin/
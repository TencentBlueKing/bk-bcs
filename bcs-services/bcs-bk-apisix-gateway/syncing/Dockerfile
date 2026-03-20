FROM golang:1.23 as builder

# Copy in the go src
WORKDIR /go/src/github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies
# RUN go mod download

# Copy the sources
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY types/ types/
COPY vendor/ vendor/

# Build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG LDFLAG

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod vendor -a -ldflags "${LDFLAG}" -o bcs-bk-apisix-gateway-syncing ./cmd/sync/main.go

FROM tencentos/tencentos4-minimal
ENV GODEBUG=x509sha1=1
WORKDIR /root
# copy swagger
COPY --from=builder /go/src/github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/bcs-bk-apisix-gateway-syncing /usr/local/bin/bcs-bk-apisix-gateway-syncing

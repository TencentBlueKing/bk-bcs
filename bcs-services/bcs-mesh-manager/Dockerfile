FROM golang:1.23 as builder

# Copy in the go src
WORKDIR /go/src/github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies
# RUN go mod download

# Copy the sources
COPY cmd/    cmd/
COPY pkg/    pkg/
COPY proto/  proto/
COPY vendor/  vendor/

# Build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod vendor -a  -o bcs-mesh-manager github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager

FROM tencentos/tencentos4-minimal
WORKDIR /root
COPY ./bin/bcs-mesh-manager /usr/local/bin/bcs-mesh-manager
COPY --from=builder /go/src/github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/bcs-mesh-manager /usr/local/bin/bcs-mesh-manager
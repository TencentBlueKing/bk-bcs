# Build the manager binary
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api api/
COPY pkg pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use cent os as base image
FROM centos:7
RUN mkdir -p /data/bcs/logs
COPY --from=builder /workspace/manager /data/bcs
# empty.tar.gz是一个包含空文件夹的压缩包， 删除某个CR时，通过上传empty到蓝鲸监控来删除监控上对应的数据
COPY config/empty.tar.gz /data/bcs

RUN chmod +x /data/bcs

WORKDIR /data/bcs
ENTRYPOINT ["/manager"]

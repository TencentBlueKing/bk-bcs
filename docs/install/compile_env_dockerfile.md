- [CentOS 8 + Golang 1.23.2](#centos-8--golang-1232)
  - [Dockerfile](#dockerfile)
  - [构建镜像命令](#构建镜像命令)
  - [测试镜像命令](#测试镜像命令)
- [CentOS 7 + Golang 1.20.2](#centos-7--golang-1202)
  - [Dockerfile](#dockerfile-1)
  - [构建镜像命令](#构建镜像命令-1)
  - [测试镜像命令](#测试镜像命令-1)
- [Ubuntu + Golang 1.23.2](#ubuntu--golang-1232)
  - [Dockerfile](#dockerfile-2)
  - [构建镜像命令](#构建镜像命令-2)
  - [测试镜像命令](#测试镜像命令-2)

##### CentOS 8 + Golang 1.23.2

###### Dockerfile
```shell
FROM centos:8

# 更新仓库并安装必要的工具和开发包
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && \
    sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && \
    sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo && \
    yum install -y wget git make vim numactl-devel && \
    dnf install -y dnf-plugins-core && \
    dnf config-manager --set-enabled powertools && \
    dnf makecache && \
    dnf install -y glibc-static && \
    yum groupinstall -y "Development Tools" && \
    yum clean all && \
    rm -rf /var/cache/yum

# 定义版本参数
ARG GOLANG_VERSION=1.23.2
ARG NODE_VERSION=20

# 下载并安装 Golang
RUN cd /tmp/ && \
    wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz

# 下载并安装 Node.js
RUN curl -sL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | bash - && \
    yum install -y nodejs && \
    yum clean all && \
    rm -rf /var/cache/yum

# 设置环境变量
ENV PATH=/usr/local/go/bin:/usr/bin:$PATH \
    GOLANG_VERSION=${GOLANG_VERSION} \
    NODE_VERSION=v${NODE_VERSION}.18.0
```

###### 构建镜像命令
```shell
docker build --build-arg GOLANG_VERSION=1.23.2 --build-arg NODE_VERSION=20 -t centos:golang-1.23.2 -f Dockerfile .
```
###### 测试镜像命令
```shell
docker run -it --rm centos:golang-1.23.2 /bin/bash -c "which go && go version && which node && node -v && which npm && npm -v"
```

##### CentOS 7 + Golang 1.20.2

###### Dockerfile
```shell
FROM centos:7

# 更新仓库和安装依赖
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && \
    sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && \
    sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo && \
    yum install -y wget git make vim numactl-devel epel-release yum-utils && \
    yum-config-manager --enable PowerTools && \
    yum install -y glibc-static && \
    yum -y groupinstall "Development Tools" && \
    yum clean all && \
    rm -rf /var/cache/yum

# 定义版本参数
ARG GOLANG_VERSION=1.20.2
ARG NODE_VERSION=16

# 安装 Golang
RUN cd /tmp/ && \
    wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz

# 安装 Node.js
RUN curl -sL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | bash - && \
    yum install -y nodejs && \
    yum clean all && \
    rm -rf /var/cache/yum

# 设置环境变量
ENV PATH=/usr/local/go/bin/:/usr/bin:$PATH \
    GOLANG_VERSION=${GOLANG_VERSION} \
    NODE_VERSION=v${NODE_VERSION}.18.0
```

###### 构建镜像命令
```shell
docker build --build-arg GOLANG_VERSION=1.20.2 --build-arg NODE_VERSION=16 -t centos:golang-1.20.2 -f Dockerfile .
```

###### 测试镜像命令
```shell
docker run -it --rm centos:golang-1.20.2 /bin/bash -c "which go && go version && which node && node -v && which npm && npm -v"
```

##### Ubuntu + Golang 1.23.2

###### Dockerfile
```shell
FROM ubuntu:latest

# 更新软件源并安装依赖
RUN apt-get update -y && apt-get install -y libnuma-dev && \
    apt-get install -y build-essential && \
    apt-get install -y wget git make vim curl

# 定义版本参数
ARG GOLANG_VERSION=1.20.2
ARG NODE_VERSION=16

# 安装Golang
RUN cd /tmp/ && \
    wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz

# 安装Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - && apt-get install -y nodejs

# 设置环境变量
ENV PATH=/usr/local/go/bin/:/usr/bin:$PATH \
    GOLANG_VERSION=${GOLANG_VERSION} \
    NODE_VERSION=v${NODE_VERSION}.18.0
```

###### 构建镜像命令
Golang 1.23.2
```shell
docker build --build-arg GOLANG_VERSION=1.23.2 --build-arg NODE_VERSION=20 -t centos:golang-1.23.2 -f Dockerfile .
```

Golang 1.20.2
```shell
docker build --build-arg GOLANG_VERSION=1.20.2 --build-arg NODE_VERSION=16 -t centos:golang-1.20.2 -f Dockerfile .
```

###### 测试镜像命令

以Golang 1.20.2版本为例

```shell
docker run -it --rm centos:golang-1.20.2 /bin/bash -c "which go && go version && which node && node -v && which npm && npm -v"
```
- [CentOS 7 本地编译](#centos-7-本地编译)
- [CentOS 8 本地编译](#centos-8-本地编译)
- [Ubuntu 本地编译](#ubuntu-本地编译)

## CentOS 7 本地编译
```shell
# 更新仓库（若需要）
sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && \
sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && \
sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

# 安装必要的工具和开发包
yum install -y wget git make vim numactl-devel epel-release yum-utils && \
yum-config-manager --enable PowerTools && \
yum install -y glibc-static && \
yum -y groupinstall "Development Tools" && \
yum clean all && \
rm -rf /var/cache/yum

# 定义版本参数
GOLANG_VERSION=1.20.2
NODE_VERSION=16

# 安装 Golang
cd /tmp/ && \
wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
rm -rf /usr/local/go && \
tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
rm -rf ./go${GOLANG_VERSION}.linux-amd64.tar.gz

# 安装 Node.js
curl -sL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | bash - && \
yum install -y nodejs && \
yum clean all && \
rm -rf /var/cache/yum

# 设置环境变量
PATH=/usr/local/go/bin/:/usr/bin:$PATH \
GOLANG_VERSION=${GOLANG_VERSION} \
NODE_VERSION=v${NODE_VERSION}.18.0
```

## CentOS 8 本地编译
```shell
# 更新仓库（若需要）
sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && \
sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && \
sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo && \

# 安装必要的工具和开发包
yum install -y wget git make vim numactl-devel && \
dnf install -y dnf-plugins-core && \
dnf config-manager --set-enabled powertools && \
dnf makecache && \
dnf install -y glibc-static && \
yum groupinstall -y "Development Tools" && \
yum clean all && \
rm -rf /var/cache/yum

# 定义版本参数
GOLANG_VERSION=1.23.2
NODE_VERSION=20

# 下载并安装 Golang
cd /tmp/ && \
wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
rm -rf /usr/local/go && \
tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
rm -rf ./go${GOLANG_VERSION}.linux-amd64.tar.gz

# 下载并安装 Node.js
curl -sL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | bash - && \
yum install -y nodejs && \
yum clean all && \
rm -rf /var/cache/yum

# 设置环境变量
PATH=/usr/local/go/bin:/usr/bin:$PATH \
GOLANG_VERSION=${GOLANG_VERSION} \
NODE_VERSION=v${NODE_VERSION}.18.0
```

## Ubuntu 本地编译
```shell
# 更新软件源并安装依赖
apt-get update -y && apt-get install -y libnuma-dev && \
apt-get install -y build-essential && \
apt-get install -y wget git make vim curl

# 定义版本参数
GOLANG_VERSION=1.20.2
NODE_VERSION=16

# 安装Golang
cd /tmp/ && \
wget https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
rm -rf /usr/local/go && \
tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
rm go${GOLANG_VERSION}.linux-amd64.tar.gz

# 安装Node.js
curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - && apt-get install -y nodejs

# 设置环境变量
PATH=/usr/local/go/bin/:/usr/bin:$PATH \
GOLANG_VERSION=${GOLANG_VERSION} \
NODE_VERSION=v${NODE_VERSION}.18.0
```
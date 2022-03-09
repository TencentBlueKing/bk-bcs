FROM alpine:3.15

ARG KUBECTL_VERSION=v1.22.0
ARG HELM_VERSION=v3.6.1

WORKDIR /root

# 配置 VIM
ENV KUBE_EDITOR=vim
COPY /root/.vimrc /root/.vimrc

RUN mkdir -p ~/.vim/colors && mkdir -p ~/.vim/autoload && \
    wget -q https://raw.githubusercontent.com/joshdick/onedark.vim/main/autoload/onedark.vim -O ~/.vim/autoload/onedark.vim && \
    wget -q https://raw.githubusercontent.com/joshdick/onedark.vim/main/colors/onedark.vim -O ~/.vim/colors/onedark.vim

# 安装依赖包
RUN apk add --update wget bash-completion vim bat

# 添加 kubectl 命令行
RUN wget -q https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    mv kubectl /usr/local/bin && \
    chmod a+x /usr/local/bin/kubectl && \
    kubectl version --client

# 添加 helm 命令行
RUN wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz  && \
    tar -xf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin && \
    rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64 && \
    helm version

# 清理缓存和不符合安全规范的命令
RUN apk del wget && \
    rm -rf /var/cache/apk/* && \
    rm -rf /sbin/apk /usr/bin/wget

# 初始化 bash 配置
RUN echo "source /etc/profile.d/bash_completion.sh" >> ~/.bashrc && \
    echo "source <(kubectl completion bash)" >> ~/.bashrc && \
    echo "export PS1='\u:\W\$ '" >> ~/.bashrc && \
    echo "export TERM=xterm-256color" >> ~/.bashrc

# 启动一个常驻进程
CMD ["/bin/sh", "-c", "sleep infinity"]

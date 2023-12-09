# Use distroless as minimal base image to package the bcs-terraform-controller binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM centos:7
# Switch to root to have permissions for operations
USER root

ARG TF_VERSION=1.3.9
ARG TARGETARCH=386

ENV TZ="Asia/Shanghai"
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

# /data/bcs/terraform/repository git代码存放路径
# /data/bcs/terraform/plugin-cache terraform插件存放路径
RUN mkdir -p /data/bcs/logs /data/bcs/terraform/repository /data/bcs/terraform/plugin-cache && chmod +x /data/bcs

RUN yum install unzip git vim gettext lrzsz net-tools telnet tcpdump wget bind-utils tree lsof -y

ADD https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_${TARGETARCH}.zip /terraform_${TF_VERSION}_linux_${TARGETARCH}.zip

RUN unzip -q /terraform_${TF_VERSION}_linux_${TARGETARCH}.zip -d /usr/local/bin/ && \
    rm /terraform_${TF_VERSION}_linux_${TARGETARCH}.zip && \
    chmod +x /usr/local/bin/terraform

WORKDIR /data/bcs

ADD ./build/bcs-terraform-controller ./

ENTRYPOINT ["/bcs-terraform-controller"]

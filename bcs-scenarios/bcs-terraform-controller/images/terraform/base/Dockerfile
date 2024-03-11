FROM centos:7

ARG TF_VERSION=1.6.0
ARG TARGETARCH=amd64

# install terraform
RUN yum install unzip git vim gettext lrzsz net-tools telnet tcpdump wget bind-utils tree lsof -y
ADD https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_${TARGETARCH}.zip /terraform_${TF_VERSION}_linux_${TARGETARCH}.zip
RUN unzip -q /terraform_${TF_VERSION}_linux_${TARGETARCH}.zip -d /usr/local/bin/ && \
    rm /terraform_${TF_VERSION}_linux_${TARGETARCH}.zip && \
    chmod +x /usr/local/bin/terraform

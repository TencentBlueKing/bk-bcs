# office centos is deprecation https://github.com/docker-library/docs/pull/2205
# use ubuntu latest lts version https://wiki.ubuntu.com/Releases
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# link sh to bash for syntax compatibility
RUN echo "dash dash/sh boolean false" | debconf-set-selections && \
    DEBIAN_FRONTEND=noninteractive dpkg-reconfigure dash

RUN export DEBIAN_FRONTEND=noninteractive && \
    sed -i 's/security\.ubuntu\.com/mirrors.cloud.tencent.com/g' /etc/apt/sources.list && \
    sed -i 's/archive\.ubuntu\.com/mirrors.cloud.tencent.com/g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y telnet curl wget && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/*

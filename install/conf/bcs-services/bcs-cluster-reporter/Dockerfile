FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo
MAINTAINER xxx xxx<xxx@tencent.com>
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" > /etc/timezone
WORKDIR /
COPY ./bcs-cluster-reporter .
COPY ./TencentSans-W7.ttf .
COPY ./TencentSans-W3.ttf .
#COPY ./conf ./conf
ENTRYPOINT ["/bcs-cluster-reporter"]

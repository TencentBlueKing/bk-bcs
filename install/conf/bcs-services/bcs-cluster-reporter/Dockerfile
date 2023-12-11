FROM centos:7
MAINTAINER xxx xxx<xxx@tencent.com>
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" > /etc/timezone
WORKDIR /
COPY ./bcs-cluster-reporter .
#COPY ./conf ./conf
ENTRYPOINT ["/bcs-cluster-reporter"]
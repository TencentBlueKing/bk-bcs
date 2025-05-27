FROM tencentos/tencentos4-minimal
MAINTAINER xxx xxx<xxx@tencent.com>
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" > /etc/timezone
WORKDIR /
COPY ./bcs-nodeagent .
#COPY ./conf ./conf
ENTRYPOINT ["/bcs-nodeagent"]

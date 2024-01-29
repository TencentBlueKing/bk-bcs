FROM alpine

RUN apk --update --no-cache add ca-certificates bash vim curl gettext libc6-compat

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-user-manager

ADD bcs-user-manager /data/bcs/bcs-user-manager/
ADD bcs-user-manager.json.template /data/bcs/bcs-user-manager/
ADD container-start.sh /data/bcs/bcs-user-manager/
RUN chmod +x /data/bcs/bcs-user-manager/container-start.sh

ENV TZ="Asia/Shanghai"
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

WORKDIR /data/bcs/bcs-user-manager/
CMD ["/data/bcs/bcs-user-manager/container-start.sh"]


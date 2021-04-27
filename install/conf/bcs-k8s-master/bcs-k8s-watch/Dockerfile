FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-watch

ADD bcs-k8s-watch /data/bcs/bcs-k8s-watch/
ADD bcs-k8s-watch.json.template /data/bcs/bcs-k8s-watch/
ADD container-start.sh /data/bcs/bcs-k8s-watch/
RUN chmod +x /data/bcs/bcs-k8s-watch/bcs-k8s-watch
RUN chmod +x /data/bcs/bcs-k8s-watch/container-start.sh

ENV TZ="Asia/Shanghai"
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime && echo ${TZ} > /etc/timezone

WORKDIR /data/bcs/bcs-k8s-watch/
CMD ["/data/bcs/bcs-k8s-watch/container-start.sh"]


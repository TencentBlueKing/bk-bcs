FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/kubeconfigs

ADD bcs-netservice-controller /data/bcs/bcs-netservice-controller/

WORKDIR /data/bcs/bcs-netservice-controller/

CMD ["/data/bcs/bcs-netservice-controller/bcs-netservice-controller"]
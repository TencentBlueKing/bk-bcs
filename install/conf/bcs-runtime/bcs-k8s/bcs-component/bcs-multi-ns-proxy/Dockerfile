FROM centos:7

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert /data/bcs/kubeconfigs

ADD bcs-multi-ns-proxy /data/bcs/bcs-multi-ns-proxy/

WORKDIR /data/bcs/bcs-multi-ns-proxy/

CMD ["/data/bcs/bcs-multi-ns-proxy/bcs-multi-ns-proxy"]
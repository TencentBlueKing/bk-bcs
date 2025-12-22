FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-k8s-watch

ADD bcs-k8s-watch /data/bcs/bcs-k8s-watch/
ADD bcs-k8s-watch.json.template /data/bcs/bcs-k8s-watch/
ADD container-start.sh /data/bcs/bcs-k8s-watch/
ADD filter.json /data/bcs/bcs-k8s-watch/
RUN chmod +x /data/bcs/bcs-k8s-watch/bcs-k8s-watch
RUN chmod +x /data/bcs/bcs-k8s-watch/container-start.sh

# 提前设置的默认env
# 关闭netservice的数据上报，默认关闭
ENV watchDisableNetService=false
# 关闭crd资源的数据上报，默认关闭
ENV watchDisableCrd=false
# 默认不添加labelSelector
ENV watchLabelSelectors="{}"

WORKDIR /data/bcs/bcs-k8s-watch/
CMD ["/data/bcs/bcs-k8s-watch/container-start.sh"]


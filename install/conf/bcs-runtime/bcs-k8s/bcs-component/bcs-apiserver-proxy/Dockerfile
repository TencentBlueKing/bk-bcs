FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs
RUN mkdir -p /data/bcs/bcs-apiserver-proxy

ADD bcs-apiserver-proxy-tools /data/bcs/bcs-apiserver-proxy/
ADD bcs-apiserver-proxy /data/bcs/bcs-apiserver-proxy/
ADD bcs-apiserver-proxy.json.template /data/bcs/bcs-apiserver-proxy/
ADD container-start.sh /data/bcs/bcs-apiserver-proxy/

RUN chmod +x /data/bcs/bcs-apiserver-proxy/bcs-apiserver-proxy
RUN chmod +x /data/bcs/bcs-apiserver-proxy/container-start.sh

WORKDIR /data/bcs/bcs-apiserver-proxy/
CMD ["/data/bcs/bcs-apiserver-proxy/container-start.sh"]

FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/istio-policy-controller

ADD istio-policy-controller /data/bcs/istio-policy-controller/
ADD container-start.sh /data/bcs/istio-policy-controller/
RUN chmod +x /data/bcs/istio-policy-controller/container-start.sh

WORKDIR /data/bcs/istio-policy-controller/
CMD ["/data/bcs/istio-policy-controller/container-start.sh"]

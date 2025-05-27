FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-ingress-controller

ADD bcs-ingress-controller /data/bcs/bcs-ingress-controller/
ADD container-start.sh /data/bcs/bcs-ingress-controller/
RUN chmod +x /data/bcs/bcs-ingress-controller/container-start.sh

WORKDIR /data/bcs/bcs-ingress-controller/
CMD ["/data/bcs/bcs-ingress-controller/container-start.sh"]

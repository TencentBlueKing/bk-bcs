FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-kube-agent

ADD bcs-kube-agent /data/bcs/bcs-kube-agent/
RUN chmod +x /data/bcs/bcs-kube-agent/bcs-kube-agent

WORKDIR /data/bcs/bcs-kube-agent/
ENTRYPOINT ["/data/bcs/bcs-kube-agent/bcs-kube-agent"]

FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-ingress-controller

ADD bcs-ingress-controller /data/bcs/bcs-ingress-controller/
ADD container-start.sh /data/bcs/bcs-ingress-controller/
RUN chmod +x /data/bcs/bcs-ingress-controller/container-start.sh

WORKDIR /data/bcs/bcs-ingress-controller/
CMD ["/data/bcs/bcs-ingress-controller/container-start.sh"]

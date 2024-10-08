# Use cent os as base image
FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo
# for command envsubset
RUN yum install -y gettext
RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-ingress-inspector

ADD bcs-ingress-inspector /data/bcs/bcs-ingress-inspector/
ADD container-start.sh /data/bcs/bcs-ingress-inspector/
RUN chmod +x /data/bcs/bcs-ingress-inspector/container-start.sh

WORKDIR /data/bcs/bcs-ingress-inspector/
CMD ["/data/bcs/bcs-ingress-inspector/container-start.sh"]

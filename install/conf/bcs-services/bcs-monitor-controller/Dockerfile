FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-monitor-controller

ADD bcs-monitor-controller /data/bcs/bcs-monitor-controller/
#ADD bcs-monitor-controller.json.template /data/bcs/bcs-monitor-controller/
ADD container-start.sh /data/bcs/bcs-monitor-controller/
RUN chmod +x /data/bcs/bcs-monitor-controller/container-start.sh

WORKDIR /data/bcs/bcs-monitor-controller/
CMD ["/data/bcs/bcs-monitor-controller/container-start.sh"]

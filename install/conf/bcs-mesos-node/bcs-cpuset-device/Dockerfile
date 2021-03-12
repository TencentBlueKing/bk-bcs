FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert

ADD bcs-cpuset-device /data/bcs/bcs-cpuset-device/
ADD container-start.sh /data/bcs/bcs-cpuset-device/
RUN chmod +x /data/bcs/bcs-cpuset-device/container-start.sh

WORKDIR /data/bcs/bcs-cpuset-device/
CMD ["/data/bcs/bcs-cpuset-device/container-start.sh"]

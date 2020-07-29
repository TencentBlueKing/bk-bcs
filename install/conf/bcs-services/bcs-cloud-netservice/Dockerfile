FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cloud-netservice/cni/conf /data/bcs/bcs-cloud-netservice/cni/bin

ADD bcs-cloud-netservice /data/bcs/bcs-cloud-netservice/
ADD container-start.sh /data/bcs/bcs-cloud-netservice/

RUN chmod +x /data/bcs/bcs-cloud-netservice/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netservice/
CMD [ "/data/bcs/bcs-cloud-netservice/container-start.sh" ]
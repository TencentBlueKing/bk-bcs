FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cloud-netcontroller/cni/conf /data/bcs/bcs-cloud-netcontroller/cni/bin

ADD bcs-cloud-netcontroller /data/bcs/bcs-cloud-netcontroller/
ADD container-start.sh /data/bcs/bcs-cloud-netcontroller/

RUN chmod +x /data/bcs/bcs-cloud-netcontroller/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netcontroller/
CMD [ "/data/bcs/bcs-cloud-netcontroller/container-start.sh" ]
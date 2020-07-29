FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cloud-netagent/cni/conf /data/bcs/bcs-cloud-netagent/cni/bin

ADD bcs-cloud-netagent /data/bcs/bcs-cloud-netagent/
ADD container-start.sh /data/bcs/bcs-cloud-netagent/

RUN chmod +x /data/bcs/bcs-cloud-netagent/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netagent/
CMD [ "/data/bcs/bcs-cloud-netagent/container-start.sh" ]

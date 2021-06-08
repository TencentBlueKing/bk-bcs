FROM centos:7

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-cloud-netagent/cni/conf /data/bcs/bcs-cloud-netagent/cni/bin

COPY bcs-cloud-netagent bcs-cloud-netagent.json.template /data/bcs/bcs-cloud-netagent/
COPY container-start.sh /data/bcs/bcs-cloud-netagent/
COPY bcs-eni-cni /data/bcs/bcs-cloud-netagent/cni/bin/
COPY bcs-eni-ipam /data/bcs/bcs-cloud-netagent/cni/bin/

RUN chmod +x /data/bcs/bcs-cloud-netagent/container-start.sh

WORKDIR /data/bcs/bcs-cloud-netagent/
CMD [ "/data/bcs/bcs-cloud-netagent/container-start.sh" ]

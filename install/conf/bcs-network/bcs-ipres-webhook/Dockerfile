FROM centos:7

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert && mkdir -p /data/bcs/bcs-ipres-webhook

COPY bcs-ipres-webhook bcs-ipres-nslabel-injector bcs-ipres-webhook.json.template /data/bcs/bcs-ipres-webhook/
COPY container-start.sh /data/bcs/bcs-ipres-webhook/

RUN chmod +x /data/bcs/bcs-ipres-webhook/container-start.sh

WORKDIR /data/bcs/bcs-ipres-webhook/
CMD [ "/data/bcs/bcs-ipres-webhook/container-start.sh" ]
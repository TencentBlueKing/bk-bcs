FROM centos:7

# for command envsubset
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs && mkdir -p /data/bcs/bcs-ipmasq-cidrsync

COPY bcs-ipmasq-cidrsync bcs-ipmasq-cidrsync.json.template /data/bcs/bcs-ipmasq-cidrsync/
COPY container-start.sh /data/bcs/bcs-ipmasq-cidrsync/

RUN chmod +x /data/bcs/bcs-ipmasq-cidrsync/container-start.sh

WORKDIR /data/bcs/bcs-ipmasq-cidrsync/
CMD [ "/data/bcs/bcs-ipmasq-cidrsync/container-start.sh" ]

FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-clusternet-controller/

ADD bcs-clusternet-controller /data/bcs/bcs-clusternet-controller/
ADD container-start.sh /data/bcs/bcs-clusternet-controller/

RUN chmod +x /data/bcs/bcs-clusternet-controller/container-start.sh

WORKDIR /data/bcs/bcs-clusternet-controller/
CMD [ "/data/bcs/bcs-clusternet-controller/container-start.sh" ]

FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert
RUN mkdir -p /data/bcs/bcs-gamestatefulset-operator/

ADD bcs-gamestatefulset-operator /data/bcs/bcs-gamestatefulset-operator/
ADD container-start.sh /data/bcs/bcs-gamestatefulset-operator/

RUN chmod +x /data/bcs/bcs-gamestatefulset-operator/bcs-gamestatefulset-operator
RUN chmod +x /data/bcs/bcs-gamestatefulset-operator/container-start.sh

WORKDIR /data/bcs/bcs-gamestatefulset-operator/
CMD [ "/data/bcs/bcs-gamestatefulset-operator/container-start.sh" ]

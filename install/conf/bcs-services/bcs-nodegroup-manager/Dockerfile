FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-nodegroup-manager/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-nodegroup-manager /data/bcs/bcs-nodegroup-manager/
ADD bcs-nodegroup-manager.json.template /data/bcs/bcs-nodegroup-manager/
ADD container-start.sh /data/bcs/bcs-nodegroup-manager/

RUN chmod +x /data/bcs/bcs-nodegroup-manager/container-start.sh
RUN chmod +x /data/bcs/bcs-nodegroup-manager/bcs-nodegroup-manager


WORKDIR /data/bcs/bcs-nodegroup-manager/
CMD [ "/data/bcs/bcs-nodegroup-manager/container-start.sh" ]

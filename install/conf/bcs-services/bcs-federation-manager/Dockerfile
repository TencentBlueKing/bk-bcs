FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-federation-manager/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-federation-manager /data/bcs/bcs-federation-manager/
ADD bcs-federation-manager.json.template /data/bcs/bcs-federation-manager/
ADD container-start.sh /data/bcs/bcs-federation-manager/

RUN chmod +x /data/bcs/bcs-federation-manager/container-start.sh
RUN chmod +x /data/bcs/bcs-federation-manager/bcs-federation-manager


WORKDIR /data/bcs/bcs-federation-manager/
CMD [ "/data/bcs/bcs-federation-manager/container-start.sh" ]

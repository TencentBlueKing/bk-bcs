FROM tencentos/tencentos4-minimal
#for command envsubst
RUN yum install -y gettext

RUN echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-image-proxy /data/bcs/bcs-image-proxy/
ADD bcs-image-proxy.json.template /data/bcs/bcs-image-proxy/
ADD container-start.sh /data/bcs/bcs-image-proxy/

RUN chmod +x /data/bcs/bcs-image-proxy/container-start.sh
RUN chmod +x /data/bcs/bcs-image-proxy/bcs-image-proxy

WORKDIR /data/bcs/bcs-image-proxy/
CMD [ "/data/bcs/bcs-image-proxy/container-start.sh" ]

FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-push-manager/
RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-push-manager /data/bcs/bcs-push-manager/
ADD bcs-push-manager.json.template /data/bcs/bcs-push-manager/
ADD container-start.sh /data/bcs/bcs-push-manager/

RUN chmod +x /data/bcs/bcs-push-manager/container-start.sh
RUN chmod +x /data/bcs/bcs-push-manager/bcs-push-manager

WORKDIR /data/bcs/bcs-push-manager/
CMD [ "/data/bcs/bcs-push-manager/container-start.sh" ]
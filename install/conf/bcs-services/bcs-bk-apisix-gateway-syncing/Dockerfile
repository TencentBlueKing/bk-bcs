FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

ADD bcs-bk-apisix-gateway-syncing /usr/local/bin/
RUN chmod +x /usr/local/bin/bcs-bk-apisix-gateway-syncing
ENV GODEBUG=x509sha1=1
WORKDIR /usr/local/bin/
CMD [ "/usr/local/bin/bcs-bk-apisix-gateway-syncing" ]

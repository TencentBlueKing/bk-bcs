FROM tencentos/tencentos4-minimal

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-mesh-manager /usr/local/bin/
ADD swagger/swagger-ui/* /data/bcs/swagger
RUN chmod +x /usr/local/bin/bcs-mesh-manager
ENV GODEBUG=x509sha1=1
WORKDIR /usr/local/bin/
CMD [ "/usr/local/bin/bcs-mesh-manager" ]

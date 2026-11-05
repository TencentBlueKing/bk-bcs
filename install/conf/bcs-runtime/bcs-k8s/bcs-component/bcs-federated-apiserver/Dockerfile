FROM tencentos/tencentos4-minimal
RUN yum install -y  ca-certificates
ENV TZ Asia/Shanghai
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
ADD apiserver /usr/bin
CMD ["/usr/bin/apiserver"]

FROM centos:7
RUN yum install -y  ca-certificates
ENV TZ Asia/Shanghai
RUN ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
ADD bcs-federated-apiserver .

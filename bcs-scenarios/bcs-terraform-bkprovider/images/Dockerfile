FROM centos:7
RUN sed -i s/mirror.centos.org/vault.centos.org/g /etc/yum.repos.d/*.repo && sed -i s/^#.*baseurl=http/baseurl=http/g /etc/yum.repos.d/*.repo && sed -i s/^mirrorlist=http/#mirrorlist=http/g /etc/yum.repos.d/*.repo

#for command envsubst
RUN yum install -y gettext

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/bcs-terraform-bkprovider/ /data/bcs/swagger
RUN mkdir -p /data/bcs/logs/bcs

COPY ./bcs-terraform-bkprovider /data/bcs/bcs-terraform-bkprovider/
COPY ./bcs-terraform-bkprovider.json.template /data/bcs/bcs-terraform-bkprovider/
COPY ./container-start.sh /data/bcs/bcs-terraform-bkprovider/
COPY ./swagger/ /data/bcs/swagger/

RUN chmod +x /data/bcs/bcs-terraform-bkprovider/container-start.sh
RUN chmod +x /data/bcs/bcs-terraform-bkprovider/bcs-terraform-bkprovider


WORKDIR /data/bcs/bcs-terraform-bkprovider/
CMD [ "/data/bcs/bcs-terraform-bkprovider/container-start.sh" ]

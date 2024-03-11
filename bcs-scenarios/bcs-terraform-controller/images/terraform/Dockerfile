FROM mirrors.tencent.com/bcs/bcs-terraform-base:latest

#for command envsubst
RUN yum install -y gettext
## yum install epel-release -y && yum install dsniff -y && yum install net-tools -y

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-terraform-controller /data/bcs/bcs-terraform-controller/
ADD bcs-terraform-controller.json.template /data/bcs/bcs-terraform-controller/
ADD container-start.sh /data/bcs/bcs-terraform-controller/
ADD .terraformrc /root/

RUN chmod +x /data/bcs/bcs-terraform-controller/container-start.sh
RUN chmod +x /data/bcs/bcs-terraform-controller/bcs-terraform-controller

WORKDIR /data/bcs/bcs-terraform-controller/
CMD [ "/data/bcs/bcs-terraform-controller/container-start.sh" ]

FROM centos:7

#for command envsubst
RUN yum install -y gettext
## yum install epel-release -y && yum install dsniff -y && yum install net-tools -y

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo 'LANG="en_US.UTF-8"' > /etc/locale.conf
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN mkdir -p /data/bcs/logs/bcs

ADD bcs-gitops-workflow /data/bcs/bcs-gitops-workflow/
ADD bcs-gitops-workflow.json.template /data/bcs/bcs-gitops-workflow/
ADD container-start.sh /data/bcs/bcs-gitops-workflow/

RUN chmod +x /data/bcs/bcs-gitops-workflow/container-start.sh
RUN chmod +x /data/bcs/bcs-gitops-workflow/bcs-gitops-workflow

WORKDIR /data/bcs/bcs-gitops-workflow/
CMD [ "/data/bcs/bcs-gitops-workflow/container-start.sh" ]

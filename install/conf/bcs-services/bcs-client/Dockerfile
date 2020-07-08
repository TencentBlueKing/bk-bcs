FROM centos:7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /var/bcs
COPY bcs.conf.template /var/bcs/bcs.conf
COPY bcs-client /usr/local/bin/

#something else

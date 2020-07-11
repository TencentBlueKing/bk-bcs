FROM centos:centos7

#for command envsubst
RUN yum install -y gettext

RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert/bcs
RUN mkdir -p /data/bcs/bcs-gateway-discovery

ADD bcs-gateway-discovery /data/bcs/bcs-gateway-discovery/
ADD bcs-gateway-discovery.json.template /data/bcs/bcs-gateway-discovery/
ADD container-start.sh /data/bcs/bcs-gateway-discovery/
ADD kong.rpm /data/bcs/bcs-gateway-discovery/
RUN chmod +x /data/bcs/bcs-gateway-discovery/container-start.sh

#install kong
RUN yum install -y perl
RUN rpm -ivh /data/bcs/bcs-gateway-discovery/kong.rpm
COPY ./bkbcs-auth /usr/local/share/lua/5.1/kong/plugins/bkbcs-auth
#kong configuration
COPY kong.conf.template /etc/kong/

WORKDIR /data/bcs/bcs-gateway-discovery/
CMD ["/data/bcs/bcs-gateway-discovery/container-start.sh"]


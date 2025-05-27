FROM tencentos/tencentos4-minimal

LABEL Author="DeveloperJim"
#all bcs modules in on container for testing

RUN yum install -y gettext
RUN yum install -y perl
RUN mkdir -p /data/bcs/logs/bcs /data/bcs/cert/bcs 

# bcs-service installation
COPY ./bcs-services/bcs-api /data/bcs/bcs-api
COPY ./bcs-services/bcs-user-manager /data/bcs/bcs-user-manager
COPY ./bcs-services/bcs-storage /data/bcs/bcs-storage
COPY ./bcs-services/bcs-client /data/bcs/bcs-client
COPY ./bcs-services/bcs-gateway-discovery /data/bcs/bcs-gateway-discovery

COPY ./bcs-services/bcs-kube-agent /data/bcs/bcs-kube-agent
COPY ./bcs-services/bcs-k8s-watch /data/bcs/bcs-k8s-watch
# bcs-k8s installation
COPY ./bcs-k8s-master/bcs-gamestatefulset-operator /data/bcs/bcs-gamestatefulset-operator

#install kong
RUN rpm -ivh /data/bcs/bcs-gateway-discovery/kong.rpm

#config type: render, sidecar, off
# > render: container start script render configuration file with ENV
# > sidecar: supported by BSCP
# > off: do nothing
ENV BCS_CONFIG_TYPE=render
ENV BCS_HOME="/data/bcs"
# bcs common
ENV log_dir="${BCS_HOME}/logs/bcs"
ENV pid_dir="/var/run/bcs"
ENV caFile="${BCS_HOME}/cert/bcs/community-ca.crt"
ENV serverCertFile="${BCS_HOME}/cert/bcs/community-server.crt"
ENV serverKeyFile="${BCS_HOME}/cert/bcs/community-server.key"
ENV clientCertFile="${BCS_HOME}/cert/bcs/community-client.crt"
ENV clientKeyFile="${BCS_HOME}/cert/bcs/community-client.key"
#!import for module discovery
ENV localIp=127.0.0.1

WORKDIR /data/bcs
